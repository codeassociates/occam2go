package modgen

import (
	"strings"
	"testing"
)

const courseSConscript = `Import('env')
local = env.Clone()

course_lib_srcs = Split('''
    utils.occ
    string.occ
    demo_cycles.occ
    demo_nets.occ
    file_in.occ
    float_io.occ
    random.occ
    ''')

shared_screen_lib_srcs = Split('''
    shared_screen.occ
    ''')
shared_screen_lib_objs = \
        [local.OccamObject(f, INCPATH='.') for f in shared_screen_lib_srcs]



course_lib = local.OccamLibrary(
        'course.lib',
        course_lib_srcs,
        INCPATH='.',
        OCCBUILDFLAGS='--include consts.inc')

def mk_shared_screen(lib_name):
    return local.OccamLibrary(
            lib_name,
            shared_screen_lib_objs,
            INCPATH='.',
            OCCBUILDFLAGS='--need course --include shared_screen.inc')

sharedscreen_lib = mk_shared_screen('shared_screen.lib')
# Build ss.lib too for backwards compatibility.
mk_shared_screen('ss.lib')
`

func TestParseSConscriptCourseSources(t *testing.T) {
	libs := ParseSConscript(courseSConscript)
	if len(libs) == 0 {
		t.Fatal("expected at least one library")
	}

	// Find course.lib
	var courseLib *Library
	for i := range libs {
		if libs[i].Name == "course.lib" {
			courseLib = &libs[i]
			break
		}
	}
	if courseLib == nil {
		t.Fatal("expected to find course.lib")
	}

	expectedSources := []string{"utils.occ", "string.occ", "demo_cycles.occ", "demo_nets.occ", "file_in.occ", "float_io.occ", "random.occ"}
	if len(courseLib.Sources) != len(expectedSources) {
		t.Fatalf("expected %d sources, got %d: %v", len(expectedSources), len(courseLib.Sources), courseLib.Sources)
	}
	for i, s := range expectedSources {
		if courseLib.Sources[i] != s {
			t.Errorf("source[%d]: expected %q, got %q", i, s, courseLib.Sources[i])
		}
	}
}

func TestParseSConscriptCourseIncludes(t *testing.T) {
	libs := ParseSConscript(courseSConscript)
	var courseLib *Library
	for i := range libs {
		if libs[i].Name == "course.lib" {
			courseLib = &libs[i]
			break
		}
	}
	if courseLib == nil {
		t.Fatal("expected to find course.lib")
	}

	if len(courseLib.Includes) != 1 || courseLib.Includes[0] != "consts.inc" {
		t.Errorf("expected includes [consts.inc], got %v", courseLib.Includes)
	}
}

func TestGenerateModuleCourse(t *testing.T) {
	libs := ParseSConscript(courseSConscript)
	var courseLib *Library
	for i := range libs {
		if libs[i].Name == "course.lib" {
			courseLib = &libs[i]
			break
		}
	}
	if courseLib == nil {
		t.Fatal("expected to find course.lib")
	}

	output := GenerateModule(*courseLib, "COURSE.MODULE")

	// Check include guard
	if !strings.Contains(output, "#IF NOT (DEFINED (COURSE.MODULE))") {
		t.Error("missing include guard #IF")
	}
	if !strings.Contains(output, "#DEFINE COURSE.MODULE") {
		t.Error("missing include guard #DEFINE")
	}
	if !strings.Contains(output, "#ENDIF") {
		t.Error("missing #ENDIF")
	}

	// Check consts.inc comes first (before source files)
	constsIdx := strings.Index(output, `#INCLUDE "consts.inc"`)
	utilsIdx := strings.Index(output, `#INCLUDE "utils.occ"`)
	if constsIdx < 0 {
		t.Error("missing consts.inc include")
	}
	if utilsIdx < 0 {
		t.Error("missing utils.occ include")
	}
	if constsIdx > utilsIdx {
		t.Error("consts.inc should come before utils.occ")
	}

	// Check all source files are included
	for _, src := range courseLib.Sources {
		if !strings.Contains(output, `#INCLUDE "`+src+`"`) {
			t.Errorf("missing #INCLUDE for %s", src)
		}
	}
}

func TestParseSplitVars(t *testing.T) {
	content := `my_srcs = Split('''
    a.occ
    b.occ
    ''')`
	vars := extractSplitVars(content)
	if files, ok := vars["my_srcs"]; !ok {
		t.Error("expected my_srcs variable")
	} else if len(files) != 2 {
		t.Errorf("expected 2 files, got %d", len(files))
	}
}

func TestGenerateModuleOutput(t *testing.T) {
	lib := Library{
		Name:     "test.lib",
		Sources:  []string{"a.occ", "b.occ"},
		Includes: []string{"consts.inc"},
	}
	output := GenerateModule(lib, "TEST.MODULE")
	expected := `#IF NOT (DEFINED (TEST.MODULE))
#DEFINE TEST.MODULE
#INCLUDE "consts.inc"
#INCLUDE "a.occ"
#INCLUDE "b.occ"
#ENDIF
`
	if output != expected {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, output)
	}
}
