package ux_test

import (
	"testing"

	"github.com/mevansam/goutils/logger"
	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestUX(t *testing.T) {
	logger.Initialize()

	RegisterFailHandler(Fail)
	RunSpecs(t, "UX")
}

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})
