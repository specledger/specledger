package spec

type FeatureInfo struct {
	RepoRoot      string   `json:"REPO_ROOT"`
	Branch        string   `json:"BRANCH"`
	FeatureDir    string   `json:"FEATURE_DIR"`
	FeatureSpec   string   `json:"FEATURE_SPEC"`
	ImplPlan      string   `json:"IMPL_PLAN"`
	Tasks         string   `json:"TASKS"`
	HasGit        bool     `json:"HAS_GIT"`
	AvailableDocs []string `json:"AVAILABLE_DOCS,omitempty"`
}

type FeatureCreateResult struct {
	BranchName string `json:"BRANCH_NAME"`
	SpecFile   string `json:"SPEC_FILE"`
	FeatureNum string `json:"FEATURE_NUM"`
}

type BranchNameConfig struct {
	Description string
	ShortName   string
	Number      int
	MaxLength   int
}

type PlanSetupResult struct {
	FeatureSpec string `json:"FEATURE_SPEC"`
	ImplPlan    string `json:"IMPL_PLAN"`
	SpecsDir    string `json:"SPECS_DIR"`
	Branch      string `json:"BRANCH"`
	HasGit      bool   `json:"HAS_GIT"`
}
