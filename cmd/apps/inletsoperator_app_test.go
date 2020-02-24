package apps

import (
	"fmt"
	"testing"
)

func Test_throwErrorForProjectIdOmission(t *testing.T) {
	projectID := ""
	providers := []string{"gce", "packet"}
	zone := "us-central1-a"

	cmd := MakeInstallInletsOperator()

	for _, provider := range providers {
		err := cmd.Flags().Set("provider", provider)
		if err != nil {
			t.Errorf("cannot set value of flag provider: %v", err)
		}
		err = cmd.Flags().Set("zone", zone)
		if err != nil {
			t.Errorf("cannot set value of flag zone: %v", err)
		}
		cmd.Flags().Set("project-id", projectID)
		if err != nil {
			t.Errorf("cannot set value of flag project-id: %v", err)
		}
		_, err = getInletsOperatorOverrides(cmd)

		if len(fmt.Sprint(err)) == 0 {
			t.Errorf("no error thrown on omitted project-id flag for provider %s", provider)
		}
	}
}
