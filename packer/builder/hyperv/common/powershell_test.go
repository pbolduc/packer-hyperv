

package common

import "testing"


func TestOutputScriptBlock(t *testing.T) {

	powershell, err := NewPowerShellv4()
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	trueOutput, err := powershell.OutputScriptBlock("$True")
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if trueOutput != "True" {
		t.Fatalf("output '%v' is not 'True'", trueOutput)
	}

	falseOutput, err := powershell.OutputScriptBlock("$False")
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if falseOutput != "False" {
		t.Fatalf("output '%v' is not 'False'", falseOutput)
	}
}

func TestRunScriptBlock(t *testing.T) {
	powershell, err := NewPowerShellv4()
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	err = powershell.RunScriptBlock("$True")
}

func TestVersion(t *testing.T) {
	powershell, err := NewPowerShellv4()
	version, err := powershell.Version();
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if (version != 4) {
		t.Fatalf("expected version 4")
	}
}
