// Copyright 2024 Matrix Origin
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

package mo_upgrade

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const (
	colorReset  = "\033[0m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorBold   = "\033[1m"
)

func printHeader(msg string) {
	fmt.Printf("\n%s%s=== %s ===%s\n\n", colorBold, colorCyan, msg, colorReset)
}

func printStep(n int, msg string) {
	fmt.Printf("\n%s[Step %d]%s %s\n", colorYellow, n, colorReset, msg)
}

func printSuccess(msg string) {
	fmt.Printf("  %s✓ %s%s\n", colorGreen, msg, colorReset)
}

func confirm(prompt string) bool {
	if autoYes {
		fmt.Printf("%s [y/N]: y (auto)\n", prompt)
		return true
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s [y/N]: ", prompt)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	return input == "y" || input == "yes"
}
