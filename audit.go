package main

import (
	"fmt"
	"slices"

	crossplanev1 "github.com/crossplane/crossplane/apis/apiextensions/v1"
)

func auditResourcesComposition(reports *[]ResourcesModeCompositionReport, comp crossplanev1.Composition) {
	fmt.Printf("*****Working with composition %v\n", comp.Name)
	resourceReports := make([]ResourceReport, 0, len(comp.Spec.Resources))

	entry := ResourcesModeCompositionReport{
		Name:            comp.Name,
		Mode:            *comp.Spec.Mode,
		ResourceReports: resourceReports,
	}

	// first check if there is a PatchSet that has the patches we want
	var patchSetMatch *crossplanev1.PatchSet = nil
	// find a PatchSet that has both ToCompositeFieldPath patches
	patchSetMatchIndex := matchingPatchSetIndex(comp.Spec.PatchSets)
	if patchSetMatchIndex > -1 {
		patchSetMatch = &comp.Spec.PatchSets[patchSetMatchIndex]
		patchSetStatusConditionsPatchIndex := slices.IndexFunc(patchSetMatch.Patches, isMatchingToCompositeFieldPathPatchFactory(StatusConditionsPath))
		patchSetStatusAtProviderPatchIndex := slices.IndexFunc(patchSetMatch.Patches, isMatchingToCompositeFieldPathPatchFactory(StatusAtProviderPath))
		entry.HasMatchingPatchSet = true
		entry.MatchingPatchSetName = patchSetMatch.Name
		entry.PatchSetStatusConditionsPatch = patchSetMatch.Patches[patchSetStatusConditionsPatchIndex]
		entry.PatchSetStatusAtProviderPatch = patchSetMatch.Patches[patchSetStatusAtProviderPatchIndex]
		// TODO
		fmt.Println("Found matching patch set", entry.MatchingPatchSetName)
	} else {
		fmt.Println("Found no matching patch set")
	}

	for i, resource := range comp.Spec.Resources {
		resourceEntry := ResourceReport{
			Name:  *resource.Name,
			Index: i,
		}
		fmt.Println("Checking resource", resourceEntry.Index, resourceEntry.Name)
		// check if the resource uses the matching PatchSet if any
		if patchSetMatch != nil && slices.ContainsFunc(resource.Patches,
			isMatchingPatchSetPatchFactory(patchSetMatch.Name)) {
			resourceEntry.UsesMatchingPatchSet = true
			fmt.Println("Resource", resourceEntry.Index, resourceEntry.Name, "uses matching patch set")
			entry.ResourceReports = append(entry.ResourceReports, resourceEntry)
			continue
		}

		// next check if it has the individual patches
		statusConditionsPatchIndex := slices.IndexFunc(resource.Patches, isMatchingToCompositeFieldPathPatchFactory(StatusConditionsPath))
		if statusConditionsPatchIndex > -1 {
			resourceEntry.StatusConditionsPatch = &resource.Patches[statusConditionsPatchIndex]
			fmt.Println("Resource", resourceEntry.Index, resourceEntry.Name, "has StatusConditionPatch defined, pointer is", resourceEntry.StatusConditionsPatch)
		} else {
			fmt.Println("Resource", resourceEntry.Index, resourceEntry.Name, "has *NO* StatusConditionPatch defined, pointer is", resourceEntry.StatusConditionsPatch)
		}

		statusAtProviderPatchIndex := slices.IndexFunc(resource.Patches, isMatchingToCompositeFieldPathPatchFactory(StatusAtProviderPath))
		if statusAtProviderPatchIndex > -1 {
			resourceEntry.StatusAtProviderPatch = &resource.Patches[statusAtProviderPatchIndex]
			fmt.Println("Resource", resourceEntry.Index, resourceEntry.Name, "has StatusAtProviderPatch defined, pointer is ", resourceEntry.StatusAtProviderPatch)
		} else {
			fmt.Println("Resource", resourceEntry.Index, resourceEntry.Name, "has *NO* StatusAtProviderPatch defined, pointer is ", resourceEntry.StatusAtProviderPatch)
		}
		resourceEntry.IsFlagged = resourceEntry.StatusConditionsPatch == nil || resourceEntry.StatusAtProviderPatch == nil
		if resourceEntry.IsFlagged {
			fmt.Println("Resource", resourceEntry.Name, "flagged")
		}
		fmt.Printf("*****Is Composition %v flagged yet? %v\n", entry.Name, entry.IsFlagged)
		entry.IsFlagged = entry.IsFlagged || resourceEntry.IsFlagged
		fmt.Printf("*****Is Composition %v flagged NOW? %v\n", entry.Name, entry.IsFlagged)
		entry.ResourceReports = append(entry.ResourceReports, resourceEntry)
		fmt.Println("Analyzed", len(entry.ResourceReports), "resources")
	}
	*reports = append(*reports, entry)
	fmt.Println("*****Analyzed", len(*reports), "compositions")
	fmt.Println("-------------------------------------------------------------------------")
}

func auditPipelineComposition(reports *[]PipelineModeCompositionReport, comp crossplanev1.Composition) (err error) {
	entry := PipelineModeCompositionReport{
		Name: comp.Name,
		Mode: *comp.Spec.Mode,
	}

	// First check to see if it uses the patch and transform function in the pipeline
	patchAndTransformFunctionIndex := slices.IndexFunc(comp.Spec.Pipeline, func(step crossplanev1.PipelineStep) bool {
		return step.FunctionRef.Name == FunctionPatchAndTransform
	})
	if patchAndTransformFunctionIndex >= 0 {
		fmt.Printf("\tComposition %s uses function-patch-and-transform\n", comp.Name)
		entry.UsesPatchAndTransform = true
	} else {
		fmt.Printf("\tComposition %s does NOT use function-patch-and-transform\n", comp.Name)
		entry.IsFlagged = true
		*reports = append(*reports, entry)
		return
	}
	*reports = append(*reports, entry)
	return
}

func matchingPatchSetIndex(patchsets []crossplanev1.PatchSet) int {
	return slices.IndexFunc(patchsets, func(ps crossplanev1.PatchSet) bool {
		containsStatusConditionsPatch := slices.ContainsFunc(ps.Patches, isMatchingToCompositeFieldPathPatchFactory(StatusConditionsPath))
		containsStatusAtProviderPatch := slices.ContainsFunc(ps.Patches, isMatchingToCompositeFieldPathPatchFactory(StatusAtProviderPath))
		return containsStatusConditionsPatch && containsStatusAtProviderPatch
	})
}
