// Copyright 2022 ADA Logics Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package planbuilder

import (
	"sync"
	"testing"

	fuzz "github.com/AdaLogics/go-fuzz-headers"

	"vitess.io/vitess/go/json2"
	"vitess.io/vitess/go/test/vschemawrapper"
	"vitess.io/vitess/go/vt/vtenv"
	"vitess.io/vitess/go/vt/sqlparser"
	"vitess.io/vitess/go/sqltypes"
	vschemapb "vitess.io/vitess/go/vt/proto/vschema"
	"vitess.io/vitess/go/vt/vtgate/vindexes"
)

var initter sync.Once

func onceInit() {
	testing.Init()
}

// loadSchemaForFuzzing is a helper to load *vindexes.VSchema
// for fuzzing.
func loadSchemaForFuzzing(f *fuzz.ConsumeFuzzer) (*vindexes.VSchema, error) {
	// formal, err := vindexes.LoadFormal(filename)
	formal, err := loadFormalForFuzzing(f)
	if err != nil {
		return nil, err
	}
	vschema := vindexes.BuildVSchema(formal, sqlparser.NewTestParser())
	if err != nil {
		return nil, err
	}
	for _, ks := range vschema.Keyspaces {
		if ks.Error != nil {
			return nil, err
		}

		for _, table := range ks.Tables {
			for i, col := range table.Columns {
				if sqltypes.IsText(col.Type) {
					table.Columns[i].CollationName = "latin1_swedish_ci"
				}
			}
		}
	}
	return vschema, nil
}

// loadFormalForFuzzing is a helper to create *vschemapb.SrvVSchema
func loadFormalForFuzzing(f *fuzz.ConsumeFuzzer) (*vschemapb.SrvVSchema, error) {
	formal := &vschemapb.SrvVSchema{}
	data, err := f.GetBytes()
	if err != nil {
		return nil, err
	}
	err = json2.UnmarshalPB(data, formal)
	if err != nil {
		return nil, err
	}
	return formal, nil
}

// FuzzTestBuilder implements the fuzzer
func FuzzTestBuilder(data []byte) int {
	initter.Do(onceInit)
	f := fuzz.NewConsumer(data)
	s, err := loadSchemaForFuzzing(f)
	if err != nil {
		return 0
	}
	env := vtenv.NewTestEnv()

	_, _ = vschemawrapper.NewVschemaWrapper(env, s, TestBuilder)
	return 1
}
