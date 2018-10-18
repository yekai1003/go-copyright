/*
处理日志订阅的问题，监控关注日志，将内容添加到数据库当中
*/
package eths

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strconv"

	"go-copyright/dbs"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

const offset = 2

func itemUnpack(index int, val interface{}, data []byte) (err error) {
	length := len(data)
	start := offset + index*32
	end := offset + (index+1)*32
	if start >= length || end >= length {
		return errors.New("too short datas")
	}
	fmt.Println("call--- itemUnpack begin", reflect.TypeOf(val).String())
	pdata := data[start:end]

	fmt.Println(string(data), string(pdata))
	if reflect.TypeOf(val).String() == "int64" || reflect.TypeOf(val).String() == "*int64" {
		var tmpval *int64 = val.(*int64)
		*tmpval, err = strconv.ParseInt(string(pdata), 16, 32)
		fmt.Println("call ParseInt", val)
	} else if reflect.TypeOf(val).String() == "string" || reflect.TypeOf(val).String() == "*string" {
		var tmpval *string = val.(*string)
		*tmpval = string(pdata)
		fmt.Println("call ParseInt", val)
	}

	fmt.Println("call--- itemUnpack end", val)
	return nil
}

func LogDataUnpack(start, end int, val interface{}, data []byte) (err error) {
	length := len(data)
	fmt.Println("call--- LogDataUnpack begin", reflect.TypeOf(val).String(), length)

	if start >= length || end > length {
		return errors.New("too short datas")
	}
	pdata := data[start:end]

	fmt.Println(string(data), string(pdata))
	if reflect.TypeOf(val).String() == "int64" || reflect.TypeOf(val).String() == "*int64" {
		var tmpval *int64 = val.(*int64)
		*tmpval, err = strconv.ParseInt(string(pdata), 16, 32)
		fmt.Println("call ParseInt", val)
	} else if reflect.TypeOf(val).String() == "string" || reflect.TypeOf(val).String() == "*string" {
		var tmpval *string = val.(*string)
		*tmpval = string(pdata)
		fmt.Println("call ParseInt", val)
	}

	fmt.Println("call--- LogDataUnpack end", val)
	return nil
}

/*
{
"address":"0x6a45c65c87ac19f594450f3eb7981b7c21d48ba4",
"topics":["0xbfb7e56ac6adbb2ff98adcfc5654c51a85c73e78cf6227e479c7d194fa142754"],
"data":"
0x
12340000000000000000000000000000
00000000000000000000000000000000
000000000000000000000000
f9a22552a978161ff51d5a259afd3fade68c26e6
71e3ece8371b6da0e6e6d69b03397767558e617f00
00000000000000000000000000000000
00000000000000000000000000000003
"
0x1234000000000000000000000000000000000000000000000000000000000000000000000000000000000000f9a22552a978161ff51d5a259afd3fade68c26e6
00000000000000000000000000000000
00000000000000000000000000000004
"blockNumber":"0x3af2e",
"transactionHash":"0x27ddfe86e738203506762deab2c399ce05571c2d6dc369b1966bed577b85b8f5",
"transactionIndex":"0x0",
"blockHash":"0xc26f3790b1db2a2819411a188e4318cecb590286f3d0cce8fa5d988c4be5c445",
"logIndex":"0x0",
"removed":false}
*/

func ParseMintEvent2Db(data []byte) error {
	fmt.Println(string(data))
	var tokenId int64
	err := LogDataUnpack(32*5, 32*6, &tokenId, data)
	if err != nil {
		fmt.Println("faile to get tokenid", err)
		return err
	}
	fmt.Println("tokenid===", tokenId)
	var pixHash string
	err = LogDataUnpack(32*0, 32*2, &pixHash, data)
	if err != nil {
		return err
	}
	fmt.Println("pixHash===", pixHash)
	var pixAddr string
	err = LogDataUnpack(88, 128, &pixAddr, data)
	if err != nil {
		return err
	}
	pixAddr = "0x" + pixAddr
	fmt.Println("pixAddr===", pixAddr)
	//插入到数据库中
	sql := fmt.Sprintf("insert into account_content(content_hash,token_id,address) values('%s',%d,'%s')", pixHash, tokenId, pixAddr)
	_, err = dbs.Create(sql)
	if err != nil {
		fmt.Println("failed to insert into mysql ", sql, err)
		return err
	}
	return nil
}

func EventSubscribe(connstr, contractAddr string) error {
	//client, err := ethclient.Dial("wss://rinkeby.infura.io/ws")
	fmt.Println("EventSubscribe() run")
	client, err := ethclient.Dial(connstr)
	if err != nil {
		//log.Fatal(err)
		return err
	}

	//合约地址转换
	contractAddress := common.HexToAddress(contractAddr)
	//填写过滤条件
	newAssetHash1 := crypto.Keccak256Hash([]byte("NewAsset(bytes32,address,uint256)"))
	query := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddress},
		Topics:    [][]common.Hash{{newAssetHash1}},
	}
	logs := make(chan types.Log)
	//订阅日志
	sub, err := client.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		//log.Fatal(err)
		return err
	}

	for {
		select {
		case err := <-sub.Err():
			fmt.Println("err", err)
			log.Fatal(err)
		case vLog := <-logs:

			switch vLog.Topics[0].Hex() {
			case newAssetHash1.Hex():
				data, err := vLog.MarshalJSON()
				fmt.Println(string(data), err)
				ParseMintEvent2Db([]byte(common.Bytes2Hex(vLog.Data)))
			}
		}
	}
	return nil
}
