package main

import crossplanev1 "github.com/crossplane/crossplane/apis/apiextensions/v1"

const ToCompositeFieldPath crossplanev1.PatchType = "ToCompositeFieldPath"
const PatchSet crossplanev1.PatchType = "PatchSet"
const StatusConditionsPath string = "status.conditions"
const StatusAtProviderPath string = "status.atProvider"
const FunctionPatchAndTransform string = "function-patch-and-transform"
