package extension

func (a *Extension) generateFingerprint() (string, error) {
	return a.api.GenerateFingerprintString(GetChromeExtensionURL() + "/#/")
}
