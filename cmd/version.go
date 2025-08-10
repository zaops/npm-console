package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "显示版本信息",
	Long:  `显示详细的版本信息，包括构建时间和 git 提交信息。`,
	RunE:  runVersion,
}

func init() {
	rootCmd.AddCommand(versionCmd)
	
	// Add flags
	versionCmd.Flags().BoolP("short", "s", false, "仅显示版本号")
	versionCmd.Flags().BoolP("json", "j", false, "以 JSON 格式输出")
}

func runVersion(cmd *cobra.Command, args []string) error {
	short, _ := cmd.Flags().GetBool("short")
	jsonOutput, _ := cmd.Flags().GetBool("json")

	if short {
		fmt.Println(Version)
		return nil
	}

	if jsonOutput {
		versionInfo := map[string]interface{}{
			"version":    Version,
			"build_time": BuildTime,
			"git_commit": GitCommit,
			"go_version": runtime.Version(),
			"platform":   fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		}
		return outputJSON(versionInfo)
	}

	// Default detailed output
	fmt.Printf("npm-console 版本 %s\n", Version)
	fmt.Printf("构建时间: %s\n", BuildTime)
	fmt.Printf("Git 提交: %s\n", GitCommit)
	fmt.Printf("Go 版本: %s\n", runtime.Version())
	fmt.Printf("平台: %s/%s\n", runtime.GOOS, runtime.GOARCH)

	return nil
}
