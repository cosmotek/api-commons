package log

import "testing"
import "github.com/rs/zerolog"

type ZerologLevelTestScenario struct {
	Input  Level
	Output zerolog.Level
}

func TestZerologLevel(t *testing.T) {
	scenarios := map[string]ZerologLevelTestScenario{
		"should default to debug": ZerologLevelTestScenario{
			Input:  "in40",
			Output: zerolog.DebugLevel,
		},
		"should return info": ZerologLevelTestScenario{
			Input:  LevelInfo,
			Output: zerolog.InfoLevel,
		},
		"should return debug": ZerologLevelTestScenario{
			Input:  LevelDebug,
			Output: zerolog.DebugLevel,
		},
		"should return warn": ZerologLevelTestScenario{
			Input:  LevelWarn,
			Output: zerolog.WarnLevel,
		},
		"should return error": ZerologLevelTestScenario{
			Input:  LevelError,
			Output: zerolog.ErrorLevel,
		},
	}

	for name, scene := range scenarios {
		t.Run(name, func(test *testing.T) {
			if out := scene.Input.ZerologLevel(); out != scene.Output {
				t.Errorf("expected '%s' but got '%s'", scene.Output, out)
			}
		})
	}
}
