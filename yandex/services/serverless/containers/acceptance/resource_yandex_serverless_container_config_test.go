package acceptance

func modeBoolToString(isReadOnly bool) string {
	if isReadOnly {
		return "ro"
	}
	return "rw"
}

func modeStringToBool(mode string) string {
	if mode == "ro" {
		return "true"
	}
	return "false"
}

type metadataOptions struct {
	awsV1HTTPEndpoint int
	gceHTTPEndpoint   int
}

type testSecretParameters struct {
	secretName   string
	secretKey    string
	secretEnvVar string
	secretValue  string
}

type testStorageMountParameters struct {
	storageMountPointName string
	storageMountPointPath string
	storageMountBucket    string
	storageMountPrefix    string
	storageMountReadOnly  bool
}

type testEphemeralDiskParameters struct {
	testMountParameters
	ephemeralDiskSizeGB      int
	ephemeralDiskBlockSizeKB int
}

type testObjectStorageParameters struct {
	testMountParameters
	objectStorageBucket string
	objectStoragePrefix string
}

type testMountParameters struct {
	mountPoint string
	mountMode  string
}

type testLogOptions struct {
	minLevel string
	disabled bool
}
