package dockerinternal

// EnvToTarget, TargetToImageName and DefaultRegistry are the single source of
// truth for the two image variants fortihugorunner knows about. pull-image,
// build-image and launch-server's --pull-latest freshness check all read from
// here so the three can't drift out of sync with each other — previously each
// command carried its own copy (launch-server's was a partial copy with a
// hardcoded registry), and a --docker-image outside that copy's list silently
// skipped the freshness check with no indication why.
var EnvToTarget = map[string]string{
	"author-dev": "prod",
	"admin-dev":  "dev",
}

var TargetToImageName = map[string]string{
	"prod": "fortinet-hugo",
	"dev":  "hugotester",
}

const DefaultRegistry = "public.ecr.aws/k4n6m5h8/"

// KnownImageName reports whether name is one of the images fortihugorunner
// can freshness-check against a registry (public ECR by default).
func KnownImageName(name string) bool {
	for _, n := range TargetToImageName {
		if n == name {
			return true
		}
	}
	return false
}
