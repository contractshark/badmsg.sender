package main

import (
	"fmt"
	"minievm/common"
	"minievm/core/vm/runtime"
)

/*
func main() {
	db, _ := ethdb.NewMemDatabase()
	state, _ := state.New(common.Hash{}, state.NewDatabase(db))

	addr := common.BytesToAddress([]byte("test1"))
	state.AddBalance(addr, big.NewInt(int64(100)))
	state.SetNonce(addr, uint64(20))
	code, _ := hex.DecodeString("60606040526001600060005055604c8060186000396000f360606040526000357c010000000000000000000000000000000000000000000000000000000090048063f8a8fd6d146039576035565b6002565b604460048050506046565b005b604a565b56")
	state.SetCode(addr, code)
	state.IntermediateRoot(false)
	fmt.Printf("code: %x\n", state.GetCode(addr))
	to := common.BytesToAddress([]byte("test2"))

	evm := vm.NewEVM(vm.Context{Transfer: core.Transfer}, state, params.TestChainConfig, vm.Config{EnableJit: false, ForceJit: false})
	value := big.NewInt(10)
	evm.StateDB.CreateAccount(to)
	evm.Transfer(evm.StateDB, addr, to, value)

	contract := vm.NewContract(vm.AccountRef(common.HexToAddress("1337")), vm.AccountRef(addr), value, uint64(100000000000))
	contract.SetCallCode(&addr, evm.StateDB.GetCodeHash(addr), evm.StateDB.GetCode(addr))

	input, _ := hex.DecodeString("")
	ret, err := vm.Run(evm, contract, input)
	fmt.Printf("ret: %x\nerr: %v\n\n", ret, err)

	fmt.Printf("StateDB:\n balance: %v\n", evm.StateDB.GetBalance(addr))
	fmt.Printf(" code: %x\n", evm.StateDB.GetCode(addr))
	return
}
*/

func main() {
	ret, _, err := runtime.Execute(common.Hex2Bytes("60606040526001600060005055604c8060186000396000f360606040526000357c010000000000000000000000000000000000000000000000000000000090048063f8a8fd6d146039576035565b6002565b604460048050506046565b005b604a565b56"), nil, nil)
	fmt.Printf("ret: %x\nerr: %v\n\n", ret, err)
}