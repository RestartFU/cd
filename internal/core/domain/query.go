package domain

type QueryDeploy struct {
	GitURL      string `query:"git_url"`
	Environment string `query:"environment"`
}
