// DO NOT EDIT: this file was generated by gen/gen_fixtures.go,
// changes made will be overwritten upon regeneration.
//
// To regenerate all fixtures, use make go_fixturegen; to regenerate only this
// test's fixtures run go generate ./pkg/relayer/miner/miner_test.go.
package miner_test

var (
	// marshaledMinableRelaysHex are the hex encoded strings of serialized
	// relayer.MinedRelays which have been pre-mined to difficulty 16 by
	// populating the signature with random bytes. It is intended for use
	// in tests.
	marshaledMinableRelaysHex = []string{
		"0a140a1212104a812fad82147600bcf3f7e29e148ec8",
		"0a140a1212102323d26a9eb27f66a22f424d469c7a2b",
		"0a140a121210bdb65e011f58209763529ed9fad5a622",
		"0a140a121210e83aae164dfe607bd06eac5d538c4339",
		"0a140a12121035f63dba4736ca57c7d796a791020e00",
	}

	// marshaledUnminableRelaysHex are the hex encoded strings of serialized
	// relayer.MinedRelays which have been pre-mined to **exclude** relays with
	// difficulty 16 (or greater). Like marshaledMinableRelaysHex, this is done
	// by populating the signature with random bytes. It is intended for use in
	// tests.
	marshaledUnminableRelaysHex = []string{
		"0a140a121210922294f2be76e0aa758f022c0950e9d2",
		"0a140a1212108b05464a92d92441d69ea625d2037877",
		"0a140a121210d1662014abe29f5ae506d260ded20431",
		"0a140a1212104464784fb01c2126b10671071dc712c1",
		"0a140a1212101e1c35adccd8197a4d31b71668cbc5ba",
	}
)
