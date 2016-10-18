package cmd

import (
	"sync"

	"github.com/spf13/cobra"

	"github.com/Shopify/themekit/kit"
)

var replaceCmd = &cobra.Command{
	Use:   "replace <filenames>",
	Short: "Overwrite theme file(s)",
	Long: `Replace will overwrite specific files if provided with file names.
If replace is not provided with file names then it will replace all
the files on shopify with your local files. Any files that do not
exist on your local machine will be removed from shopify.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := initializeConfig(cmd.Name(), true); err != nil {
			return err
		}

		wg := sync.WaitGroup{}
		for _, client := range themeClients {
			wg.Add(1)
			go replace(client, args, &wg)
		}
		wg.Wait()
		return nil
	},
}

func replace(client kit.ThemeClient, filenames []string, wg *sync.WaitGroup) error {
	jobQueue := client.Process(wg)
	defer close(jobQueue)

	assetsActions := map[string]kit.AssetEvent{}
	if len(filenames) == 0 {
		assets, err := client.AssetList()
		if err != nil {
			return err
		}

		for _, asset := range assets {
			assetsActions[asset.Key] = kit.NewRemovalEvent(asset)
		}

		localAssets, err := client.LocalAssets()
		if err != nil {
			return err
		}

		for _, asset := range localAssets {
			assetsActions[asset.Key] = kit.NewUploadEvent(asset)
		}
	} else {
		for _, filename := range filenames {
			asset, err := client.LocalAsset(filename)
			if err != nil {
				return err
			} else if asset.IsValid() {
				assetsActions[asset.Key] = kit.NewUploadEvent(asset)
			}
		}
	}
	for _, event := range assetsActions {
		jobQueue <- event
	}
	return nil
}
