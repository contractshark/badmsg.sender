package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"minievm/common"
	"minievm/detectors"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"

	"net/http"
	_ "net/http/pprof"

	_ "github.com/lib/pq"
)

type Contract struct {
	Abi string `json:"abi"`
	Bin string `json:"bin"`
}
type SolcOutput struct {
	Contracts map[string]Contract `json:"contracts"`
	Version   string              `json:"version"`
}

func RunOverflowDetector(contractName, codeHex, abi, functionName string, initAttacker, initVictims bool) {
	ofd := &detectors.OverFlowDetector{}
	ofd.Init(contractName, abi)
	ofd.OverFlowDetector(codeHex, functionName, initAttacker, initVictims)
	if ofd.Detected {
		fmt.Printf("\n\tAddr: %s\n\tReason: %s\n\tFunction: %s\n\tInput: %02x\n", contractName, ofd.Reason, functionName, ofd.Unpackedinput)
	}
}

func RunDetector(code []byte) {
	var solcOutput SolcOutput
	functionNames := []string{"transfer", "transferFrom"}
	json.Unmarshal(code, &solcOutput)
	for contractName, contract := range solcOutput.Contracts {
		if strings.Contains(contractName, "itMaps.sol") {
			continue
		}
		for _, fname := range functionNames {
			for i := 0; i < 4; i++ {
				RunOverflowDetector(contractName, contract.Bin, contract.Abi, fname, (i>>1)&1 == 0, i&1 == 0)
			}
		}
	}
}

func CallFunction(code []byte, functionName string) {
	var solcOutput SolcOutput
	json.Unmarshal(code, &solcOutput)
	for contractName, contract := range solcOutput.Contracts {
		if strings.Contains(contractName, "itMaps.sol") {
			continue
		}
		ofd := &detectors.OverFlowDetector{}
		ofd.Init(contractName, contract.Abi)
		_, ok := ofd.ABI.Methods[functionName]
		if ok {
			code := common.Hex2Bytes(contract.Bin)
			ofd.CreateContract(ofd.Owner, code, big.NewInt(0))
			// ofd.SetCallback(detectors.EVMCallback)
			var key common.Hash
			ofd.SetCallback(func(args []interface{}) {
				loc := args[0].(common.Hash)
				val := args[1].(*big.Int)
				log.Printf("%02X, %02X\n", loc, val)
				key = loc
			})
			ofd.CallFunctionStub(functionName)
			log.Printf("Get State: %02X\n", ofd.GetState(key))
		}
	}
}

func main2() {
	filePath := flag.String("p", "./", "path to file or folder")
	callee := flag.String("call", "", "Call specific function w/o fuzzing")
	// scanDir := flag.Bool("b", false, "batch scan")
	flag.Parse()

	//Common Channel for the goroutines
	tasks := make(chan *exec.Cmd, 16)
	var wg sync.WaitGroup
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			for cmd := range tasks {
				stdoutStderr, err := cmd.Output()
				if err != nil {
					// log.Println("Oops...", cmd.Args[2], " ", err)
				} else {
					if *callee == "" {
						RunDetector(stdoutStderr)
					} else {
						CallFunction(stdoutStderr, *callee)
					}
				}
			}
			wg.Done()
		}()
	}

	// Fire tasks
	cmd := "solc"
	solcArgs := []string{"--combined-json=bin,abi"}

	contractPath := *filePath
	fi, err := os.Stat(contractPath)
	if err != nil {
		fmt.Println(err)
		return
	}
	switch mode := fi.Mode(); {
	case mode.IsDir():
		files, err := ioutil.ReadDir(contractPath)
		if err != nil {
			log.Fatal(err)
		}

		// Exclude scanned result
		for _, f := range files {
			log.Println(f.Name())
			if strings.Contains(f.Name(), "itMaps.sol") {
				continue
			}
			tasks <- exec.Command(cmd, append(solcArgs, contractPath+f.Name())...)
		}
	case mode.IsRegular():
		tasks <- exec.Command(cmd, append(solcArgs, contractPath)...)
	}

	close(tasks)

	// wait for the workers to finish
	wg.Wait()

	return
}

func main() {
	go http.ListenAndServe(":8080", http.DefaultServeMux)
	contractPath := flag.String("p", "./", "path to file or folder")
	logPath := flag.String("lp", "./fuzz_log", "fuzzer's log path")
	solcPath := flag.String("sp", "solc", "solc path")
	flag.Parse()

	dispatcher(*solcPath, *contractPath, *logPath)
}

func dispatcher(solcpath, contractpath, logpath string) {
	tasks := make(chan *detectors.FuzzInt, 16)
	var wg sync.WaitGroup
	for i := 0; i < 1; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range tasks {

				task.FuzzContracts()
			}
		}()
	}

	fi, err := os.Stat(contractpath)
	if err != nil {
		fmt.Println(err)
		return
	}
	switch mode := fi.Mode(); {
	case mode.IsDir():
		files, err := ioutil.ReadDir(contractpath)
		if err != nil {
			log.Fatal(err)
		}

		// count := len(files)
		// bar := pb.StartNew(count)

		for _, f := range files {
			// bar.Increment()
			log.Printf("Now Fuzzing... %s\n", f.Name())
			tasks <- detectors.NewContractFuzzer(solcpath, path.Join(contractpath, f.Name()), logpath, false)
			// runtime.GC()
			// common.PrintMemUsage()
		}
	case mode.IsRegular():
		tasks <- detectors.NewContractFuzzer(solcpath, contractpath, logpath, true)
	}

	wg.Wait()
	close(tasks)
}
