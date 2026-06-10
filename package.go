package botsframework

// Main code for the package is in the `botsfw` directory.

// packageName returns the name of this root package.
//
// It exists only so this otherwise statement-less package contributes a
// (trivially) covered statement to coverage reports — the real code lives in the
// botsfw subpackage.
func packageName() string {
	return "botsframework"
}
