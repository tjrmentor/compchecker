package main

import (
	"encoding/json"
	"os"
	"testing"

	crossplanev1 "github.com/crossplane/crossplane/apis/apiextensions/v1"
)

func TestAuditResourcesComposition(t *testing.T) {
	var testCompsBytes, _ = os.ReadFile("compsTest.json")
	var testComps crossplanev1.CompositionList
	json.Unmarshal(testCompsBytes, &testComps)

	t.Run("Unpatched Resources Composition", func(t *testing.T) {
		testComp := testComps.Items[0]
		testReport := []ResourcesModeCompositionReport{}
		auditResourcesComposition(&testReport, testComp)
		testEntry := testReport[0]
		// fmt.Println(testEntry)
		if !testEntry.IsFlagged {
			t.Fatalf("entry was not flagged")
		}
	})

	t.Run("Patched by patch set Resources Composition", func(t *testing.T) {
		testComp := testComps.Items[1]
		testReport := []ResourcesModeCompositionReport{}
		auditResourcesComposition(&testReport, testComp)
		testEntry := testReport[0]
		// fmt.Println(testEntry)
		if testEntry.IsFlagged {
			t.Fatalf("entry was flagged")
		}
	})

}

func TestAuditPipelineComposition(t *testing.T) {
	var testCompsBytes, _ = os.ReadFile("compsTest.json")
	var testComps crossplanev1.CompositionList
	json.Unmarshal(testCompsBytes, &testComps)

	t.Run("Should flag composition that doesn't used patch and transform", func(t *testing.T) {
		testComp := testComps.Items[3]
		testReport := []PipelineModeCompositionReport{}
		auditPipelineComposition(&testReport, testComp)
		testEntry := testReport[0]

		if !testEntry.IsFlagged || testEntry.UsesPatchAndTransform {
			t.Fatalf("Got IsFlagged: %v, UsesPatchAndTransform: %v, wanted (true, false)",
				testEntry.IsFlagged, testEntry.UsesPatchAndTransform)
		}
	})

}
func TestIsMatchingToCompositeFieldPathPatchFactory(t *testing.T) {
	var testCompsBytes, _ = os.ReadFile("compsTest.json")
	var testComps crossplanev1.CompositionList
	json.Unmarshal(testCompsBytes, &testComps)

	statusConditionsMatcher := isMatchingToCompositeFieldPathPatchFactory(StatusConditionsPath)
	testCompOne := testComps.Items[1]

	for i, patch := range testCompOne.Spec.PatchSets[0].Patches {
		// fmt.Println("***** Patch", i, "for TestCompOne")
		// fmt.Println(patchString(patch))
		isMatch := statusConditionsMatcher(patch)
		if i == 2 && !isMatch {
			t.Fatalf("TestCompOne: Did not match index 2 as expected")
		}
		if i != 2 && isMatch {
			t.Fatalf("TestCompOne: Matched index %d, expected no match", i)
		}
	}

	testCompTwo := testComps.Items[2]
	for _, patch := range testCompTwo.Spec.PatchSets[0].Patches {
		// fmt.Println("***** Patch", i, "for TestCompTwo")
		// fmt.Println(patchString(patch))
		isMatch := statusConditionsMatcher(patch)
		if isMatch {
			t.Fatalf("TestCompTwo: Found match, no match expected")
		}
	}

}

func TestMatchingPatchSetIndex(t *testing.T) {
	var testCompsBytes, _ = os.ReadFile("compsTest.json")
	var testComps crossplanev1.CompositionList
	json.Unmarshal(testCompsBytes, &testComps)

	t.Run("Should return negative index when no patch set field", func(t *testing.T) {
		noPatchSetComp := testComps.Items[0]
		expected := -1
		actual := matchingPatchSetIndex(noPatchSetComp.Spec.PatchSets)
		if expected != actual {
			t.Fatalf("Expected %d, got %d\n", expected, actual)
		}
	})
	t.Run("Should return non-negative index", func(t *testing.T) {
		firstPatchSetComp := testComps.Items[1]
		expected := 0
		actual := matchingPatchSetIndex(firstPatchSetComp.Spec.PatchSets)
		if expected != actual {
			t.Fatalf("For first comp, expected %d, got %d\n", expected, actual)
		}
	})

	t.Run("Should return negative index when no match", func(t *testing.T) {
		secondPatchSetComp := testComps.Items[2]
		expected := -1
		actual := matchingPatchSetIndex(secondPatchSetComp.Spec.PatchSets)
		if expected != actual {
			t.Fatalf("Expected %d, got %d\n", expected, actual)
		}
	})
}
