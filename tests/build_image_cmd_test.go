package dockerinternal_test

import (
        "fmt"
	"bytes"
        //"context"
	//"errors"
	"testing"

	"docker-run-go/cmd"
	//"docker-run-go/dockerinternal"
	"docker-run-go/mockdocker"
	"github.com/stretchr/testify/assert"
)

func TestBuildImageCmd(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		mockClient   *mockdocker.MockDockerClient
		expectErr    bool
		expectOutput string
	}{
		{
			name:         "success with author-dev",
			args:         []string{"author-dev"},
			mockClient:   &mockdocker.MockDockerClient{},
			expectErr:    false,
			expectOutput: "**** Built a author-dev container named: fortinet-hugo ****",
		},
		{
			name:         "invalid environment arg",
			args:         []string{"wrong-env"},
			mockClient:   &mockdocker.MockDockerClient{},
			expectErr:    true,
			expectOutput: "Usage: docker-run-go build-image [author-dev | admin-dev]",
		},
		{
			name:         "image pull fails",
			args:         []string{"admin-dev"},
			mockClient:   &mockdocker.MockDockerClient{FailPull: true},
			expectErr:    true,
			expectOutput: "Couldn't pull required image, exiting....",
		},
		{
			name:         "build fails",
			args:         []string{"admin-dev"},
			mockClient:   &mockdocker.MockDockerClient{FailBuild: true},
			expectErr:    true,
			expectOutput: "failed to build image",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
                        if tc.mockClient == nil {
                                t.Fatal("mockClient is nil in test case: " + tc.name)
                        }
			buf := new(bytes.Buffer)

			cmd := cmd.NewTestableBuildImageCmd(tc.mockClient)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs(tc.args)
			err := cmd.Execute()

			output := buf.String()
                        fmt.Printf("Output of test function: \n%s", output)
                        fmt.Printf("tc.expectOutput: \n%s\n", tc.expectOutput)
			if tc.expectErr {
				assert.Error(t, err)
				assert.Contains(t, output, tc.expectOutput)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, output, tc.expectOutput)
			}
		})
	}
}
