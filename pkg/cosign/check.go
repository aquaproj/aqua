package cosign

import "os/exec"

func (verifier *Verifier) HasCosign() bool {
	_, err := exec.LookPath("cosign")
	return err == nil
}
