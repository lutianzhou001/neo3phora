package api

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/joeqian10/neo3-gogogo/crypto"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"go.mongodb.org/mongo-driver/bson"

	"neo3fura_http/lib/type/h160"
	"neo3fura_http/var/stderr"
)

func (me *T) GetBridgeTxByNonce(args struct {
	ContractHash h160.T
	TokenHash    h160.T
	Nonce        int64
	Limit        int64
	Skip         int64
	Filter       map[string]interface{}
}, ret *json.RawMessage) error {
	var filter bson.M
	nonceStr := strconv.FormatInt(args.Nonce, 10)
	if args.ContractHash.Valid() == false {
		return stderr.ErrInvalidArgs
	}

	filter = bson.M{"contract": args.ContractHash.Val(),
		"$or":                 []interface{}{bson.M{"eventname": "GasWithdrawal"}, bson.M{"eventname": "NativeClaimable"}},
		"state.value.0.value": nonceStr,
	}

	if args.TokenHash.Valid() == true && args.TokenHash.Val() != "0xd2a4cff31913016155e38e474a2c06d08be276cf" {
		token, _ := util.Uint160DecodeStringLE(strings.TrimPrefix(args.TokenHash.Val(), "0x"))

		encoded := crypto.Base64Encode(token.BytesBE())
		filter = bson.M{"contract": args.ContractHash.Val(),
			"$or":                 []interface{}{bson.M{"eventname": "TokenWithdrawal"}, bson.M{"eventname": "TokenClaimable"}},
			"state.value.1.value": nonceStr,
			"state.value.0.value": encoded,
		}
	}

	r1, _, err := me.Client.QueryAll(struct {
		Collection string
		Index      string
		Sort       bson.M
		Filter     bson.M
		Query      []string
		Limit      int64
		Skip       int64
	}{
		Collection: "Notification",
		Index:      "GetBridgeTxByNonce",
		Sort:       bson.M{"_id": -1},
		Filter:     filter,
		Query:      []string{},
		Limit:      args.Limit,
		Skip:       args.Skip,
	}, ret)

	var result map[string]interface{}
	if len(r1) > 0 {
		result = r1[0]
	}
	r, err := json.Marshal(result)
	if err != nil {
		return err
	}
	*ret = json.RawMessage(r)

	return nil
}
