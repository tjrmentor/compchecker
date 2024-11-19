package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"text/template"

	crossplanev1 "github.com/crossplane/crossplane/apis/apiextensions/v1"
)

type ResourceReport struct {
	Name                  string
	Index                 int
	UsesMatchingPatchSet  bool
	StatusConditionsPatch *crossplanev1.Patch
	StatusAtProviderPatch *crossplanev1.Patch
	IsFlagged             bool
}

func (rpt ResourcesModeCompositionReport) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Name: %v\n", rpt.Name))
	sb.WriteString(fmt.Sprintf("Mode: %v\n", rpt.Mode))
	sb.WriteString(fmt.Sprintf("IsFlagged: %v\n", rpt.IsFlagged))
	sb.WriteString(fmt.Sprintf("HasMatchingPatchSet: %v\n", rpt.HasMatchingPatchSet))
	if rpt.HasMatchingPatchSet {
		sb.WriteString(fmt.Sprintf("MatchingPatchSetName: %v\n", rpt.MatchingPatchSetName))
		sb.WriteString(fmt.Sprintf("PatchSetStatusConditionsPatch:\n %v", patchString(rpt.PatchSetStatusConditionsPatch)))
		sb.WriteString(fmt.Sprintf("PatchSetStatusAtProviderPatch:\n %v", patchString(rpt.PatchSetStatusAtProviderPatch)))
	}
	return sb.String()
}

type ResourcesModeCompositionReport struct {
	Name                          string
	Mode                          crossplanev1.CompositionMode
	IsFlagged                     bool
	ResourceReports               []ResourceReport
	HasMatchingPatchSet           bool
	MatchingPatchSetName          string
	PatchSetStatusConditionsPatch crossplanev1.Patch
	PatchSetStatusAtProviderPatch crossplanev1.Patch
}

type PipelineModeCompositionReport struct {
	Name                  string
	Mode                  crossplanev1.CompositionMode
	IsFlagged             bool
	UsesPatchAndTransform bool
	ResourcesInputReports []ResourceReport
}

func patchString(p crossplanev1.Patch) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\t\tType: %v\n", p.Type))
	if p.Type == PatchSet {
		sb.WriteString(fmt.Sprintf("\t\tPatchSetName: %v\n", *p.PatchSetName))
	} else {
		sb.WriteString(fmt.Sprintf("\t\tFromFieldPath: %v\n", *p.FromFieldPath))
		sb.WriteString(fmt.Sprintf("\t\tToFieldPath: %v\n", *p.ToFieldPath))
	}

	return sb.String()
}

func report(resourcesModeReports []ResourcesModeCompositionReport, pipelineModeReports []PipelineModeCompositionReport) (err error) {
	fmt.Println("There are ", len(resourcesModeReports), "resource mode compostitions")
	resourcesCompositionsTemplate := `RESOURCE MODE COMPOSITIONS
Index	Name	HasMatchingPatchSet	MatchingPatchSetName	IsFlagged
{{ range $index, $element := . -}}
{{ $index }}	{{ $element.Name }}	{{ $element.HasMatchingPatchSet }}	{{ $element.MatchingPatchSetName }}	{{ $element.IsFlagged }}
{{ end}}`
	reportTmpl := template.Must(template.New("").Parse(resourcesCompositionsTemplate))
	outfile, _ := os.Create("resourceCompositions.txt")
	w := tabwriter.NewWriter(outfile, 5, 4, 3, '\t', 0)
	if err = reportTmpl.Execute(w, resourcesModeReports); err != nil {
		return
	}
	w.Flush()

	resourcesTemplate := `RESOURCE MODE CRs BY COMPOSITION
Parent Index	Parent Comp	Resource Name	Resource Index	UsesMatchingPatchSet	IsFlagged
{{ range $parentIndex, $parent := . }}{{ if len $parent.ResourceReports }}{{ range $parent.ResourceReports }}{{ $parentIndex }}	{{ $parent.Name }}	{{ .Index }}	{{ .Name }}	{{ .UsesMatchingPatchSet }}	{{ .IsFlagged }}
{{ end }}{{ else }}{{ $parentIndex}}	{{ $parent.Name }}	N/A	N/A	N/A	N/A
{{ end }}{{ end }}`
	reportTmpl = template.Must(template.New("").Parse(resourcesTemplate))
	outfile, _ = os.Create("resources.txt")
	w = tabwriter.NewWriter(outfile, 5, 4, 3, '\t', 0)
	if err = reportTmpl.Execute(w, resourcesModeReports); err != nil {
		return
	}
	w.Flush()

	// 	pipelineTmpl := `Name	UsesPatchAndTransform	IsFlagged
	// {{ range . }}
	// 		{{- .Name }}	{{ .UsesPatchAndTransform }}	{{ .IsFlagged }}
	// {{ end }}`
	// 	reportTmpl = template.Must(template.New("").Parse(pipelineTmpl))
	// 	outfile, _ = os.Create("pipeline.txt")
	// 	w = tabwriter.NewWriter(outfile, 5, 4, 3, '\t', 0)
	// 	if err = reportTmpl.Execute(w, pipelineModeReports); err != nil {
	// 		return
	// 	}
	// 	w.Flush()

	return nil
}
