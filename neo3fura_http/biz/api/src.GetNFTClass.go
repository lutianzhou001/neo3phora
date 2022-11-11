package api

import (
	"encoding/base64"
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"neo3fura_http/lib/type/NFTstate"
	"neo3fura_http/lib/type/h160"
	"neo3fura_http/var/stderr"
	"strconv"
	"strings"
	"time"
)

func (me *T) GetNFTClass(args struct {
	MarketHash h160.T
	AssetHash  h160.T
	NFTSate    string
	Filter     map[string]interface{}
	Raw        *map[string]interface{}
}, ret *json.RawMessage) error {
	if args.AssetHash.Valid() == false {
		return stderr.ErrInvalidArgs
	}

	if args.MarketHash.Valid() == false {
		return stderr.ErrInvalidArgs
	}
	//currentTime := time.Now().UnixNano() / 1e6
	//result := make([]map[string]interface{}, 0)

	//var filter bson.M
	//if args.NFTSate == NFTstate.Auction.Val(){
	//	//filter = bson.M{"market": args.MarketHash,"asset": args.AssetHash,"auctionType":2,"amount":1}
	//	filter = bson.M{"market":args.MarketHash,"asset":args.AssetHash,"eventname":"Auction","$or": []interface{}{
	//		bson.M{"extendData": bson.M{"$regex": "auctionType\":\"2", "$options": "$i"}},
	//		bson.M{"extendData": bson.M{"$regex": "auctionType\": \" 2", "$options": "$i"}},
	//	}}
	//}else if args.NFTSate == NFTstate.Sale.Val(){
	//	filter = bson.M{"market":args.MarketHash,"asset":args.AssetHash,"eventname":"Auction","$or": []interface{}{
	//		bson.M{"extendData": bson.M{"$regex": "auctionType\":\"1", "$options": "$i"}},
	//		bson.M{"extendData": bson.M{"$regex": "auctionType\": \" 1", "$options": "$i"}},
	//	}}
	//}else{
	//	filter = bson.M{"market":args.MarketHash,"asset":args.AssetHash,"eventname":"Auction",}
	//}

	//var r1, err = me.Client.QueryAggregate(
	//	struct {
	//		Collection string
	//		Index      string
	//		Sort       bson.M
	//		Filter     bson.M
	//		Pipeline   []bson.M
	//		Query      []string
	//	}{
	//		Collection: "MarketNotification",
	//		Index:      "GetNFTClass",
	//		Sort:       bson.M{},
	//		Filter:     bson.M{},
	//		Pipeline: []bson.M{
	//			bson.M{"$match": filter},
	//			bson.M{"$lookup": bson.M{
	//				"from": "SelfControlNep11Properties",
	//				"let":  bson.M{"asset": "$asset", "tokenid": "$tokenid"},
	//				"pipeline": []bson.M{
	//					bson.M{"$match": bson.M{"$expr": bson.M{"$and": []interface{}{
	//						bson.M{"$eq": []interface{}{"$tokenid", "$$tokenid"}},
	//						bson.M{"$eq": []interface{}{"$asset", "$$asset"}},
	//					}}}},
	//					bson.M{"$set": bson.M{"class": "$image"}},
	//				//	bson.M{"$sort"}
	//				},
	//				"as": "properties"},
	//			},
	//			bson.M{"$lookup": bson.M{
	//				"from": "Market",
	//				"let":  bson.M{"asset": "$asset", "tokenid": "$tokenid"},
	//				"pipeline": []bson.M{
	//					bson.M{"$match":bson.M{"amount":1}},
	//					bson.M{"$match": bson.M{"$expr": bson.M{"$and": []interface{}{
	//						bson.M{"$eq": []interface{}{"$tokenid", "$$tokenid"}},
	//						bson.M{"$eq": []interface{}{"$asset", "$$asset"}},
	//					}}}},
	//				},
	//				"as": "marketInfo"},
	//			},
	//			bson.M{"$group": bson.M{"_id": bson.M{"asset": "$asset", "class": "$properties.class"}, "asset": bson.M{"$last": "$asset"}, "tokenid": bson.M{"$last": "$tokenid"}, "properties": bson.M{"$last": "$properties"}, "marketInfo": bson.M{"$last": "$marketInfo"},
	//				  "extendData": bson.M{"$last": "$extendData"}, "deadline": bson.M{"$last": "$extendData"},"infoArr": bson.M{"$push": "$$ROOT"}}},
	//			//bson.M{"$project": bson.M{"_id": 1, "properties": 1, "asset": 1, "tokenid": 1, "propertiesArr": 1, "auctionAmount": 1, "deadline": 1, "timestamp": 1}},
	//			//bson.M{"$group": bson.M{"_id": bson.M{"asset": "$asset", "class": "$properties.class"}, "asset": bson.M{"$last": "$asset"}, "tokenid": bson.M{"$last": "$tokenid"},"properties": bson.M{"$last": "$properties"}}},
	//			//bson.M{"$project": bson.M{"_id": 1, "properties": 1, "asset": 1, "tokenid": 1}},
	//
	//		},
	//		Query: []string{},
	//	}, ret)
	//
	//if err != nil {
	//	return err
	//}

	var filter bson.M
	if args.NFTSate == NFTstate.Auction.Val() {
		filter = bson.M{"market": args.MarketHash, "amount": 1, "auctionType": 2}
	} else if args.NFTSate == NFTstate.Sale.Val() {
		filter = bson.M{"market": args.MarketHash, "amount": 1, "auctionType": 1}
	} else {
		filter = bson.M{"amount": 1}
	}

	var r2, err = me.Client.QueryAggregate(
		struct {
			Collection string
			Index      string
			Sort       bson.M
			Filter     bson.M
			Pipeline   []bson.M
			Query      []string
		}{
			Collection: "SelfControlNep11Properties",
			Index:      "GetNFTClass",
			Sort:       bson.M{},
			Filter:     bson.M{},
			Pipeline: []bson.M{
				bson.M{"$match": bson.M{"asset": args.AssetHash}},
				bson.M{"$lookup": bson.M{
					"from": "Market",
					"let":  bson.M{"asset": "$asset", "tokenid": "$tokenid"},
					"pipeline": []bson.M{
						bson.M{"$match": filter},
						bson.M{"$match": bson.M{"$expr": bson.M{"$and": []interface{}{
							bson.M{"$eq": []interface{}{"$tokenid", "$$tokenid"}},
							bson.M{"$eq": []interface{}{"$asset", "$$asset"}},
						}}}},
						//	bson.M{"$sort"}
					},
					"as": "marketInfo"},
				},
				bson.M{"$set": bson.M{"class": "$image"}},
				bson.M{"$group": bson.M{"_id": bson.M{"asset": "$asset", "class": "$class"}, "class": bson.M{"$last": "$class"}, "asset": bson.M{"$last": "$asset"}, "tokenid": bson.M{"$last": "$tokenid"},
					"name": bson.M{"$last": "$name"}, "image": bson.M{"$last": "$image"}, "supply": bson.M{"$last": "$supply"}, "thumbnail": bson.M{"$last": "$thumbnail"},
					"properties": bson.M{"$last": "$properties"}, "marketArr": bson.M{"$last": "$marketInfo"}, "itemList": bson.M{"$push": "$$ROOT"}}},
			},
			Query: []string{},
		}, ret)

	if err != nil {
		return err
	}

	//result := make([]map[string]interface{},0)
	for _, item := range r2 {
		marketInfo := item["marketArr"].(primitive.A)[0].(map[string]interface{})
		marketInfo = GetNFTState(marketInfo, args.MarketHash)
		item["currentBidAmount"] = marketInfo["bidAmount"]
		item["currentBidAsset"] = marketInfo["auctionAsset"]
		//item["state"] =marketInfo["state"]
		item["auctionType"] = marketInfo["auctionType"]
		item["auctionAsset"] = marketInfo["auctionAsset"]
		item["auctionAmount"] = marketInfo["auctionAmount"]
		item["deadline"] = marketInfo["deadline"]

		count := 0
		groupinfo := item["itemList"].(primitive.A)
		for _, it := range groupinfo {
			pit := it.(map[string]interface{})
			market := pit["marketInfo"].(primitive.A)[0].(map[string]interface{})
			market = GetNFTState(market, args.MarketHash)
			if market["state"].(string) == "sale" || market["state"].(string) == "auction" {
				count++
			}
		}
		item["claimed"] = len(groupinfo) - count

		asset := item["asset"].(string)
		image := item["image"]
		if image != nil {
			item["image"] = ImagUrl(asset, item["image"].(string), "images")
		} else {
			item["image"] = ""
		}
		if item["thumbnail"] != nil {
			tb, err2 := base64.URLEncoding.DecodeString(item["thumbnail"].(string))
			if err2 != nil {
				return err2
			}
			item["thumbnail"] = ImagUrl(item["asset"].(string), string(tb[:]), "thumbnail")

		} else {
			item["thumbnail"] = ImagUrl(item["asset"].(string), image.(string), "thumbnail")
		}
		if item["name"] != nil {
			item["name"] = item["name"]
		} else {
			item["name"] = ""
		}
		if item["number"] != nil {
			item["number"] = item["number"]
		} else {
			strArray := strings.Split(item["name"].(string), "#")
			if len(strArray) >= 2 {
				number := strArray[1]
				n, err22 := strconv.ParseInt(number, 10, 64)
				if err22 != nil {
					item["number"] = int64(-1)
				}
				item["number"] = n
			} else {
				item["number"] = int64(-1)
			}
		}

		if item["supply"] != nil {
			series, err2 := base64.URLEncoding.DecodeString(item["supply"].(string))
			if err2 != nil {
				return err2
			}
			item["supply"] = string(series)
		} else {
			item["supply"] = ""
		}

		if item["video"] != nil {
			item["video"] = item["video"]
		} else {
			item["video"] = ""
		}

		delete(item, "_id")
		delete(item, "itemList")
		delete(item, "marketArr")
		delete(item, "properties")
		delete(item, "class")
		//result = append(result, item)

	}

	//mapsort.MapSort5(result, "number")

	//count := len(result)

	r3, err := me.FilterAggragateAndAppendCount(r2, len(r2), args.Filter)

	if err != nil {
		return err
	}
	r, err := json.Marshal(r3)
	if err != nil {
		return err
	}
	if args.Raw != nil {
		*args.Raw = r3
	}
	*ret = json.RawMessage(r)
	return nil
}

func GetNFTState(info map[string]interface{}, primarymarket interface{}) map[string]interface{} {
	if len(info) > 0 {
		deadline := info["deadline"].(int64)
		auctionType := info["auctionType"].(int32)
		bidAmount := info["bidAmount"].(primitive.Decimal128).String()
		market := info["market"]
		info["currentBidAmount"] = info["bidAmount"]
		info["currentBidAmount"] = info["auctionAsset"]
		currentTime := time.Now().UnixNano() / 1e6
		if deadline > currentTime && market == primarymarket {
			if auctionType == 1 {
				info["state"] = "sale" //
			} else if auctionType == 2 {
				info["state"] = "auction"
			}
		} else if deadline <= currentTime && market == primarymarket {
			if auctionType == 2 && bidAmount != "0" {
				info["state"] = "soldout" //竞拍有人出价
			} else {
				info["state"] = "expired"
			}
		} else {
			info["state"] = "soldout"
		}

	} else {
		info["state"] = ""
	}

	delete(info, "bidAmount")
	delete(info, "bidder")
	delete(info, "auctor")
	delete(info, "timestamp")
	return info
}
