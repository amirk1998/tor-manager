package tor

// BuiltInBridges contains a curated set of built-in bridges derived from
// the Tor Browser source. These are updated periodically by the Tor Project.
//
// Note: Built-in bridges are public and may be blocked in some regions.
// For higher censorship environments, request fresh bridges via BridgeDB.
var BuiltInBridges = map[TransportType][]string{
	TransportObfs4: {
		"obfs4 217.23.14.26:443 D9A82D2F9C2F65A18407B1D2B764F130847F8B5D cert=BJE0tIiy7Cqu5XAuVnZdA8yt2b1Nv8DpFJByNW4ODWHFUNy4/jRSFUBnRCOF45o0yFqeJg iat-mode=0",
		"obfs4 38.229.1.78:80  C8CBDB2464FC9804A69531437BCF2BE31FDD2EE4 cert=Hmyfd2ev46gGY7NoVxkn1WWIEr56rNQQfkQr2RVYkDMlPZq3hHe9uBb4Txp3lRn4bsGrYg iat-mode=1",
		"obfs4 85.31.186.98:443 011F2599C0E9B27EE74B353155E244813763C3E5 cert=ayq0XzCwhpdysn5o0EyDUbmSOx3X/oTEbzDMvK8sB8WQ+E2mQWBHQdOXaMbMt9n0ATyNhg iat-mode=0",
		"obfs4 85.31.186.26:443 91A6354697E6B02A386312F68D82CF86D1C41CE9 cert=bAfIityHn3goJmBzBBGBwXGQxoiO2+dxQJQm9DZGO/XNe1T9x4IyuaqFvqFyXSxFVLXRNA iat-mode=0",
		"obfs4 193.11.166.194:27015 2D82C2E354D531A68469ADF7F878FA88F1B33B5D cert=4TLQPJrTSaDffMK7Nbao6LC7G9OcE+MkJy+znVDeVYoRODNueSt7tcOXyqYG1UKEEgjzXA iat-mode=0",
		"obfs4 193.11.166.194:27025 1AE2C08904527FEA90C4C4F8C1083EA59FBC6FAF cert=f+unsYj/sGTJvbUFaYttVMGnIn9WkRfNFKkP/OWSRGGO5hNkn6jteN+E0otq0Fo2CXHZBA iat-mode=0",
	},
	TransportSnowflake: {
		"snowflake 192.0.2.3:1 2B280B23E1107BB62ABFC40DDCC8824814F80A72 fingerprint=2B280B23E1107BB62ABFC40DDCC8824814F80A72 url=https://snowflake-broker.torproject.net/ front=cdn.sstatic.net ice=stun:stun.l.google.com:19302,stun:stun.antisip.com:3478,stun:stun.bluesip.net:3478,stun:stun.dus.net:3478 utls-imitate=hellorandomizedalpn",
		"snowflake 192.0.2.4:1 8838024498816A039FCBBAB14E6F40A0843051FA fingerprint=8838024498816A039FCBBAB14E6F40A0843051FA url=https://snowflake-broker.torproject.net/ front=cdn.sstatic.net ice=stun:stun.l.google.com:19302,stun:stun.antisip.com:3478,stun:stun.bluesip.net:3478,stun:stun.dus.net:3478 utls-imitate=hellorandomizedalpn",
	},
	TransportMeekAzure: {
		"meek_lite 0.0.2.0:2 B9E7141C594AF25699E0079C1F0146F409495296 url=https://meek.azureedge.net/ front=ajax.aspnetcdn.com",
	},
	TransportWebTunnel: {
		"webtunnel 45.33.1.189:443 54BF1146B161573185898CD60A9AE6BF484F776A url=https://webtunnel.torproject.org:443/vTAbJGLaEoqELAGM",
	},
}

// BuiltInBridgeEntries returns all built-in bridges for a given transport,
// parsed into Bridge structs. Invalid lines are silently skipped.
func BuiltInBridgeEntries(tt TransportType) []Bridge {
	lines, ok := BuiltInBridges[tt]
	if !ok {
		return nil
	}
	var result []Bridge
	for _, line := range lines {
		b, err := ParseBridgeLine(line)
		if err == nil {
			result = append(result, b)
		}
	}
	return result
}

// AllBuiltInTransports returns the list of transport types that have built-in bridges.
func AllBuiltInTransports() []TransportType {
	return []TransportType{
		TransportObfs4,
		TransportSnowflake,
		TransportMeekAzure,
		TransportWebTunnel,
	}
}
