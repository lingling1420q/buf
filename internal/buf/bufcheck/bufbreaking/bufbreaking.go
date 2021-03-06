// Copyright 2020 Buf Technologies, Inc.
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

// Package bufbreaking contains the breaking change detection functionality.
//
// The primary entry point to this package is the Handler.
package bufbreaking

import (
	"context"

	"github.com/bufbuild/buf/internal/buf/bufanalysis"
	"github.com/bufbuild/buf/internal/buf/bufcheck"
	"github.com/bufbuild/buf/internal/buf/bufcheck/bufbreaking/internal/bufbreakingv1beta1"
	"github.com/bufbuild/buf/internal/buf/bufcheck/internal"
	"github.com/bufbuild/buf/internal/buf/bufcore/bufimage"
	"go.uber.org/zap"
)

// Handler handles the main breaking functionality.
type Handler interface {
	// Check runs the breaking checks.
	//
	// The image should have source code info for this to work properly. The previousImage
	// does not need to have source code info.
	//
	// Images should be filtered with regards to imports before passing to this function.
	Check(
		ctx context.Context,
		config *Config,
		previousImage bufimage.Image,
		image bufimage.Image,
	) ([]bufanalysis.FileAnnotation, error)
}

// NewHandler returns a new Handler.
func NewHandler(logger *zap.Logger) Handler {
	return newHandler(logger)
}

// Checker is a checker.
type Checker interface {
	bufcheck.Checker

	internalBreaking() *internal.Checker
}

// Config is the check config.
type Config struct {
	// Checkers are the checkers to run.
	//
	// Checkers will be sorted by first categories, then id when Configs are
	// created from this package, i.e. created wth ConfigBuilder.NewConfig.
	Checkers               []Checker
	IgnoreIDToRootPaths    map[string]map[string]struct{}
	IgnoreRootPaths        map[string]struct{}
	IgnoreUnstablePackages bool
}

// GetCheckers returns the checkers.
//
// Should only be used for printing.
func (c *Config) GetCheckers() []bufcheck.Checker {
	return checkersToBufcheckCheckers(c.Checkers)
}

// NewConfigV1Beta1 returns a new Config.
func NewConfigV1Beta1(externalConfig ExternalConfigV1Beta1) (*Config, error) {
	internalConfig, err := internal.ConfigBuilder{
		Use:                           externalConfig.Use,
		Except:                        externalConfig.Except,
		IgnoreRootPaths:               externalConfig.Ignore,
		IgnoreIDOrCategoryToRootPaths: externalConfig.IgnoreOnly,
		IgnoreUnstablePackages:        externalConfig.IgnoreUnstablePackages,
	}.NewConfig(
		bufbreakingv1beta1.VersionSpec,
	)
	if err != nil {
		return nil, err
	}
	return internalConfigToConfig(internalConfig), nil
}

// GetAllCheckersV1Beta1 gets all known checkers.
//
// Should only be used for printing.
func GetAllCheckersV1Beta1() ([]bufcheck.Checker, error) {
	config, err := NewConfigV1Beta1(
		ExternalConfigV1Beta1{
			Use: bufbreakingv1beta1.VersionSpec.AllCategories,
		},
	)
	if err != nil {
		return nil, err
	}
	return checkersToBufcheckCheckers(config.Checkers), nil
}

// ExternalConfigV1Beta1 is an external config.
type ExternalConfigV1Beta1 struct {
	Use    []string `json:"use,omitempty" yaml:"use,omitempty"`
	Except []string `json:"except,omitempty" yaml:"except,omitempty"`
	// IgnoreRootPaths
	Ignore []string `json:"ignore,omitempty" yaml:"ignore,omitempty"`
	// IgnoreIDOrCategoryToRootPaths
	IgnoreOnly             map[string][]string `json:"ignore_only,omitempty" yaml:"ignore_only,omitempty"`
	IgnoreUnstablePackages bool                `json:"ignore_unstable_packages,omitempty" yaml:"ignore_unstable_packages,omitempty"`
}

func internalConfigToConfig(internalConfig *internal.Config) *Config {
	return &Config{
		Checkers:               internalCheckersToCheckers(internalConfig.Checkers),
		IgnoreIDToRootPaths:    internalConfig.IgnoreIDToRootPaths,
		IgnoreRootPaths:        internalConfig.IgnoreRootPaths,
		IgnoreUnstablePackages: internalConfig.IgnoreUnstablePackages,
	}
}

func configToInternalConfig(config *Config) *internal.Config {
	return &internal.Config{
		Checkers:               checkersToInternalCheckers(config.Checkers),
		IgnoreIDToRootPaths:    config.IgnoreIDToRootPaths,
		IgnoreRootPaths:        config.IgnoreRootPaths,
		IgnoreUnstablePackages: config.IgnoreUnstablePackages,
	}
}

func checkersToBufcheckCheckers(checkers []Checker) []bufcheck.Checker {
	if checkers == nil {
		return nil
	}
	s := make([]bufcheck.Checker, len(checkers))
	for i, e := range checkers {
		s[i] = e
	}
	return s
}
