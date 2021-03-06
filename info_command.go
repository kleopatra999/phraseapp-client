package main

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/phrase/phraseapp-go/phraseapp"
)

type ClientInfo struct {
	BuiltAt                 string `json:"built_at"`
	ClientRevision          string `json:"client_revision"`
	LibraryRevision         string `json:"library_revision"`
	ClientDocRevision       string `json:"client_doc_revision"`
	DocRevision             string `json:"doc_revision"`
	ClientRevisionGenerator string `json:"client_generator_revision"`
	RevisionGenerator       string `json:"generator_revision"`
	GoVersion               string `json:"go_version"`
}

func NewInfo() ClientInfo {
	return ClientInfo{
		BuiltAt:                 BUILT_AT,
		ClientRevision:          REVISION,
		LibraryRevision:         LIBRARY_REVISION,
		ClientDocRevision:       RevisionDocs,
		DocRevision:             phraseapp.RevisionDocs,
		ClientRevisionGenerator: RevisionGenerator,
		RevisionGenerator:       phraseapp.RevisionGenerator,
		GoVersion:               runtime.Version(),
	}
}

func GetInfo() string {
	info := []string{
		fmt.Sprintf("Built at:                            %s", BUILT_AT),
		fmt.Sprintf("PhraseApp Client version:            %s", PHRASEAPP_CLIENT_VERSION),
		fmt.Sprintf("PhraseApp Client revision:           %s", REVISION),
		fmt.Sprintf("PhraseApp Library revision:          %s", LIBRARY_REVISION),
		fmt.Sprintf("PhraseApp Docs revision client:      %s", RevisionDocs),
		fmt.Sprintf("PhraseApp Docs revision lib:         %s", phraseapp.RevisionDocs),
		fmt.Sprintf("PhraseApp Generator revision client: %s", RevisionGenerator),
		fmt.Sprintf("PhraseApp Generator revision lib:    %s", phraseapp.RevisionGenerator),
		fmt.Sprintf("GoVersion:                           %s", runtime.Version()),
	}
	return fmt.Sprintf("%s\n", strings.Join(info, "\n"))
}

func infoCommand() error {
	fmt.Print(GetInfo())
	return nil
}

var (
	REVISION                 = "DEV"
	LIBRARY_REVISION         = "DEV"
	BUILT_AT                 = "LIVE"
	PHRASEAPP_CLIENT_VERSION = "DEV"
)
