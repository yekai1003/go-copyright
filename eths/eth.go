package eths

import (
	"fmt"
	"go-copyright/configs"
	"go-copyright/utils"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

func NewAcc(pass string) (string, error) {
	// Connect the client
	client, err := rpc.Dial(configs.Config.Eth.ConnStr)
	if err != nil {
		//log.Fatalf("could not create ipc client: %v", err)
		fmt.Println("failed to connect geth,", err)
		return "", err
	}

	var account string
	err = client.Call(&account, "personal_newAccount", pass)

	if err != nil {
		fmt.Println("failed to newAccount ", err)
		return "", err
	}

	// Print events from the subscription as they arrive.
	fmt.Printf("account: %s\n", account)
	return account, nil
}

//mint(bytes32 _hash, uint256 _price, uint256 _weight, string _data)
func UploadPic(from, pass, hash, data string) (err error) {
	//1. 客户端连接
	client, err := ethclient.Dial(configs.Config.Eth.ConnStr)
	if err != nil {
		fmt.Println("failed to connect geth ", err)
		return err
	}
	//2. 合约入口
	pxaInstance, err := NewPxa(common.HexToAddress(configs.Config.Eth.PxaAddr), client)
	if err != nil {
		fmt.Println("failed to NewPxa", err)
		return err
	}
	fmt.Println(pxaInstance)
	//	auth := &bind.TransactOpts{
	//		From:     common.HexToAddress(from),
	//		GasLimit: 300000,
	//	}

	//3. 设置签名 -- 需要owner的keystore文件
	filename, err := utils.GetFileName(string([]rune(from)[2:]), configs.Config.Eth.Keydir)
	if err != nil || filename == "" {
		fmt.Println("faile to get filename by address", err, configs.Config.Eth.Keydir, from)
		return err
	}

	f, err := os.Open(configs.Config.Eth.Keydir + "/" + filename)
	if err != nil {
		fmt.Println("faile to Open filename", err, filename)
		return err
	}

	defer f.Close()
	auth, err := bind.NewTransactor(f, pass)
	if err != nil {
		fmt.Println("faile to NewTransactor", err, configs.Config.Eth.Keydir, from)
		return err
	}
	fmt.Printf("signature:%+v\n", auth)
	//挖矿
	_, err = pxaInstance.Mint(auth, common.HexToHash(hash), big.NewInt(100), big.NewInt(100), data)
	if err != nil {
		fmt.Println("failed to call mint ", err)
		return err
	}
	return err
}

//资产分割
func EthSplitAsset(tokenID, weight int64, buyer string) error {
	//1. 客户端连接
	client, err := ethclient.Dial(configs.Config.Eth.ConnStr)
	if err != nil {
		fmt.Println("failed to connect geth ", err)
		return err
	}
	//2. 合约入口
	pxaInstance, err := NewPxa(common.HexToAddress(configs.Config.Eth.PxaAddr), client)
	if err != nil {
		fmt.Println("failed to NewPxa", err)
		return err
	}
	fmt.Println(pxaInstance)

	//3. 设置签名 -- 需要owner的keystore文件
	filename, err := utils.GetFileName(string([]rune(configs.Config.Eth.Foundation)[2:]), configs.Config.Eth.Keydir)
	if err != nil || filename == "" {
		fmt.Println("faile to get filename by address", err, configs.Config.Eth.Keydir, configs.Config.Eth.Foundation)
		return err
	}

	f, err := os.Open(configs.Config.Eth.Keydir + "/" + filename)
	if err != nil {
		fmt.Println("faile to Open filename", err, filename)
		return err
	}

	defer f.Close()
	auth, err := bind.NewTransactor(f, "found")
	if err != nil {
		fmt.Println("faile to NewTransactor", err, configs.Config.Eth.Keydir, configs.Config.Eth.Foundation)
		return err
	}
	fmt.Printf("signature:%+v\n", auth)
	//4. 分割资产
	//SplitAsset(opts *bind.TransactOpts, _tokenId *big.Int, _weight *big.Int, _buyer common.Address)
	_, err = pxaInstance.SplitAsset(auth, big.NewInt(tokenID), big.NewInt(weight), common.HexToAddress(buyer))
	if err != nil {
		fmt.Println("failed to call SplitAsset ", err)
		return err
	}
	return err
}

//转账
func EthTransfer20(_from, _pass, _to string, _price int64) error {
	//1. 客户端连接
	client, err := ethclient.Dial(configs.Config.Eth.ConnStr)
	if err != nil {
		fmt.Println("failed to connect geth ", err)
		return err
	}
	//2. 合约入口
	pxcInstance, err := NewPxc(common.HexToAddress(configs.Config.Eth.PxcAddr), client)
	if err != nil {
		fmt.Println("failed to NewPxa", err)
		return err
	}
	//fmt.Println(pxaInstance)

	//3. 设置签名 -- 需要owner的keystore文件
	filename, err := utils.GetFileName(string([]rune(_from)[2:]), configs.Config.Eth.Keydir)
	if err != nil || filename == "" {
		fmt.Println("faile to get filename by address", err, configs.Config.Eth.Keydir, _from)
		return err
	}

	f, err := os.Open(configs.Config.Eth.Keydir + "/" + filename)
	if err != nil {
		fmt.Println("faile to Open filename", err, filename)
		return err
	}

	defer f.Close()
	auth, err := bind.NewTransactor(f, _pass)
	if err != nil {
		fmt.Println("faile to NewTransactor", err, configs.Config.Eth.Keydir, _from)
		return err
	}
	fmt.Printf("signature:%+v\n", auth)
	//4. 转账 Transfer(opts *bind.TransactOpts, _to common.Address, _value *big.Int) 买家(auth)->卖家()
	_, err = pxcInstance.Transfer(auth, common.HexToAddress(_to), big.NewInt(_price))
	if err != nil {
		fmt.Println("failed to call SplitAsset ", err)
		return err
	}
	return nil
}
