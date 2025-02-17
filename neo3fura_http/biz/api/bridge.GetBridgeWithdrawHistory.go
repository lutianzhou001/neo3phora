package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"net/http"
	"os"
	"strings"

	"github.com/joeqian10/neo3-gogogo/crypto"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"neo3fura_http/lib/neox"
	"neo3fura_http/lib/type/h160"
	"neo3fura_http/var/stderr"
)

func (me *T) GetBridgeDepositHistory(args struct {
	ContractHash h160.T
	Sender       h160.T
	Limit        int64
	Skip         int64
	Filter       map[string]interface{}
}, ret *json.RawMessage) error {
	if args.ContractHash.Valid() == false {
		return stderr.ErrInvalidArgs
	}
	sender, _ := util.Uint160DecodeStringLE(strings.TrimPrefix(args.Sender.Val(), "0x"))
	encoded := crypto.Base64Encode(sender.BytesBE())
	var filter bson.M
	if args.Sender.Valid() == true {
		filter = bson.M{
			"contract": args.ContractHash.Val(),
			"$or":      []interface{}{bson.M{"eventname": "GasDeposit", "state.value.3.value": encoded}, bson.M{"eventname": "NativeDeposit", "state.value.3.value": encoded}, bson.M{"eventname": "TokenDeposit", "state.value.5.value": encoded}},
		}
	} else {
		filter = bson.M{
			"contract": args.ContractHash.Val(),
			"$or":      []interface{}{bson.M{"eventname": "NativeDeposit"}, bson.M{"eventname": "GasDeposit"}, bson.M{"eventname": "TokenDeposit"}},
		}
	}

	r1, count, err := me.Client.QueryAll(struct {
		Collection string
		Index      string
		Sort       bson.M
		Filter     bson.M
		Query      []string
		Limit      int64
		Skip       int64
	}{
		Collection: "Notification",
		Index:      "GetBridgeDepositHistory",
		Sort:       bson.M{"timestamp": -1},
		Filter:     filter,
		Query:      []string{},
		Limit:      args.Limit,
		Skip:       args.Skip,
	}, ret)

	//get status of target chain
	for _, item := range r1 {
		eventname := item["eventname"].(string)
		if eventname == "GasDeposit" || eventname == "NativeDeposit" {
			value := item["state"].(map[string]interface{})["value"].(primitive.A)
			nonce := value[0].(map[string]interface{})["value"]
			to := value[1].(map[string]interface{})["value"].(string)
			toDecode, err := base64.StdEncoding.DecodeString(to)
			if err != nil {
				return fmt.Errorf("fail to decode to toAddress: %w", err)
			}
			toAddress, err := util.Uint160DecodeBytesLE(toDecode)
			if err != nil {
				return fmt.Errorf("fail to Uint160DecodeBytesLE for toAddress: %w", err)
			}

			amountFromStore := value[2].(map[string]interface{})["value"].(string)
			amount := interfaceToBigInt(amountFromStore)

			from := value[3].(map[string]interface{})["value"].(string)
			fromDecode, err := base64.StdEncoding.DecodeString(from)
			if err != nil {
				return fmt.Errorf("fail to decode to fromAddress: %w", err)
			}
			fromAddress, err := util.Uint160DecodeBytesLE(fromDecode)
			if err != nil {
				return fmt.Errorf("fail to Uint160DecodeBytesLE for fromAddress: %w", err)
			}

			item["from"] = "0x" + fromAddress.String()
			item["to"] = "0x" + toAddress.String()
			item["amount"] = BigIntToDecimalData(amount, 8)
			item["symbol"] = "GAS"
			item["nonce"] = nonce

			tx, status, err := getDepositTxFromNeox("", nonce.(string))
			if err != nil {
				return err
			}
			item["neoxTx"] = tx
			item["status"] = status

		} else {
			value := item["state"].(map[string]interface{})["value"].(primitive.A)
			n3Token := value[0].(map[string]interface{})["value"].(string)
			n3TokenStr, err := base64.StdEncoding.DecodeString(n3Token)
			if err != nil {
				return fmt.Errorf("fail to decode to neoxToken: %w", err)
			}
			n3TokenDecode, err := util.Uint160DecodeBytesLE(n3TokenStr)
			if err != nil {
				return fmt.Errorf("fail to Uint160DecodeBytesLE for neoxToken: %w", err)
			}

			neoxToken := value[1].(map[string]interface{})["value"].(string)
			neoxTokenStr, err := base64.StdEncoding.DecodeString(neoxToken)
			if err != nil {
				return fmt.Errorf("fail to decode to neoxToken: %w", err)
			}
			neoxTokenDecode, err := util.Uint160DecodeBytesLE(neoxTokenStr)
			if err != nil {
				return fmt.Errorf("fail to Uint160DecodeBytesLE for neoxToken: %w", err)
			}

			nonce := value[2].(map[string]interface{})["value"]
			to := value[3].(map[string]interface{})["value"].(string)
			toDecode, err := base64.StdEncoding.DecodeString(to)
			if err != nil {
				return fmt.Errorf("fail to decode to toAddress: %w", err)
			}
			toAddress, err := util.Uint160DecodeBytesLE(toDecode)
			if err != nil {
				return fmt.Errorf("fail to Uint160DecodeBytesLE for toAddress: %w", err)
			}

			amountFromStore := value[4].(map[string]interface{})["value"]
			amount := interfaceToBigInt(amountFromStore)
			from := value[5].(map[string]interface{})["value"].(string)
			fromDecode, err := base64.StdEncoding.DecodeString(from)
			if err != nil {
				return fmt.Errorf("fail to decode to fromAddress: %w", err)
			}
			fromAddress, err := util.Uint160DecodeBytesLE(fromDecode)
			if err != nil {
				return fmt.Errorf("fail to Uint160DecodeBytesLE for fromAddress: %w", err)
			}

			r2, _, err := me.Client.QueryAll(struct {
				Collection string
				Index      string
				Sort       bson.M
				Filter     bson.M
				Query      []string
				Limit      int64
				Skip       int64
			}{
				Collection: "Asset",
				Index:      "Asset",
				Sort:       bson.M{"_id": -1},
				Filter:     bson.M{"hash": "0x" + n3TokenDecode.String()},
				Query:      []string{},
			}, ret)

			if err != nil {
				return err
			}

			symbol := r2[0]["symbol"]
			decimal := r2[0]["decimals"].(int32)

			item["from"] = "0x" + fromAddress.String()
			item["to"] = "0x" + toAddress.String()

			item["amount"] = BigIntToDecimalData(amount, int64(decimal))
			item["symbol"] = symbol
			item["nonce"] = nonce
			item["neoxToken"] = "0x" + neoxTokenDecode.String()

			tx, status, err := getDepositTxFromNeox("0x"+neoxTokenDecode.String(), nonce.(string))
			if err != nil {
				return err
			}
			item["neoxTx"] = tx

			item["status"] = status
		}

		delete(item, "state")
		delete(item, "Vmstate")
		delete(item, "index")
		delete(item, "_id")

	}

	r2, err := me.FilterArrayAndAppendCount(r1, count, args.Filter)
	if err != nil {
		return err
	}
	r, err := json.Marshal(r2)
	if err != nil {
		return err
	}
	*ret = json.RawMessage(r)

	return nil
}

func getDepositTxFromNeox(tokenHash string, nonceStr string) (string, string, error) {
	rt := os.ExpandEnv("${RUNTIME}")
	var url string
	switch rt {
	case "staging":
		url = neox.MainBridgeDepositTxUrl
	case "test":
		url = neox.TestNetBridgeDepositTxUrl
	default:
		url = neox.TestNetBridgeDepositTxUrl
	}
	var urlStr string
	if tokenHash == "" {
		urlStr = url + nonceStr
	} else {
		urlStr = url + tokenHash + "/" + nonceStr
	}

	resp, err := http.Get(urlStr)
	if err != nil {
		return "", "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	result := make(map[string]interface{})
	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", "", nil
	}
	if result["txid"] == nil {
		return "", "pending", nil
	}
	txid := result["txid"].(string)
	status := "success"
	return txid, status, err
}

func interfaceToBigInt(amount interface{}) *big.Int {
	if amount == nil {
		return big.NewInt(0)
	}
	amountStr := fmt.Sprintf("%v", amount)
	bi := new(big.Int)
	_, ok := bi.SetString(amountStr, 10)
	if !ok {
		return big.NewInt(0)
	} else {
		return bi
	}
}

func BigIntToDecimalData(amount *big.Int, decimal int64) string {
	divisor := new(big.Int)
	divisor.Exp(big.NewInt(10), big.NewInt(decimal), nil)

	//result := new(big.Int)
	//result.Div(num, divisor)

	numRat := new(big.Rat).SetInt(amount)
	divisorRat := new(big.Rat).SetInt(divisor)

	result := new(big.Rat).Quo(numRat, divisorRat)

	return strings.TrimRight(strings.TrimRight(result.FloatString(6), "0"), ".")

}

func (me *T) getTokenInfo(token string, ret *json.RawMessage) (string, int64, error) {
	r1, _, err := me.Client.QueryAll(struct {
		Collection string
		Index      string
		Sort       bson.M
		Filter     bson.M
		Query      []string
		Limit      int64
		Skip       int64
	}{
		Collection: "Asset",
		Index:      "Asset",
		Sort:       bson.M{"_id": -1},
		Filter:     bson.M{"hash": token},
		Query:      []string{},
	}, ret)

	if err != nil {
		return "", 0, err
	}

	symbol := r1[0]["symbol"]
	decimal := r1[0]["decimals"]
	return symbol.(string), decimal.(int64), nil
}
