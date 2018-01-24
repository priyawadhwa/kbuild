package main

import (
	"fmt"
	"github.com/containers/image/copy"
	"github.com/containers/image/signature"
	"github.com/containers/image/transports/alltransports"

	"os"
)

func main() {
	srcDirPath := "docker-archive:/Users/priyawadhwa/go/src/github.com/priyawadhwa/kbuild/appender/testdata/python.tar"
	destImg := "docker://gcr.io/priya-wadhwa/upload:test"

	srcRef, err := alltransports.ParseImageName(srcDirPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	destRef, err := alltransports.ParseImageName(destImg)
	if err != nil {
		fmt.Println("failed to get destRef: ", err)
		os.Exit(1)
	}

	policyContext, err := getPolicyContext()
	if err != nil {
		fmt.Println("failed to get policy context: ", err)
		os.Exit(1)
	}

	err = copy.Image(policyContext, destRef, srcRef, nil)

	if err != nil {
		fmt.Println("failed to copy image: ", err)
		os.Exit(1)
	}

}

func getPolicyContext() (*signature.PolicyContext, error) {
	policy, err := signature.DefaultPolicy(nil)
	if err != nil {
		fmt.Println("Error retrieving policy")
		return nil, err
	}

	policyContext, err := signature.NewPolicyContext(policy)
	if err != nil {
		fmt.Println("Error retrieving policy context")
		return nil, err
	}
	return policyContext, nil
}
