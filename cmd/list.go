package cmd

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"

	"github.com/YaleSpinup/spinup-cli/pkg/spinup"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

var listSpaceCost bool
var showFailedResources bool

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.AddCommand(listSpacesCmd)
	listCmd.AddCommand(listResourcesCmd)
	listCmd.AddCommand(listImagesCmd)

	listSpacesCmd.PersistentFlags().BoolVarP(&listSpaceCost, "cost", "c", false, "Query for the space cost")
	listResourcesCmd.PersistentFlags().BoolVar(&showFailedResources, "show-failed", false, "Also show failed resources")
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List spinup objects",
}

var listSpacesCmd = &cobra.Command{
	Use:   "spaces",
	Short: "List spaces",
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debug("Listing Spaces")

		spaces := spinup.Spaces{}
		if err := SpinupClient.GetResource(map[string]string{}, &spaces); err != nil {
			return err
		}

		if listSpaceCost {
			for _, s := range spaces.Spaces {
				spaceCost := &spinup.SpaceCost{}
				if err := SpinupClient.GetResource(map[string]string{"id": s.Id.String()}, spaceCost); err != nil {
					return err
				}

				s.Cost = spaceCost
			}
		}

		j, err := json.MarshalIndent(spaces.Spaces, "", "  ")
		if err != nil {
			return err
		}

		f := bufio.NewWriter(os.Stdout)
		defer f.Flush()
		f.Write(j)

		return nil
	},
}

var listResourcesCmd = &cobra.Command{
	Use:   "resources",
	Short: "Lists the resources in your space(s)",
	RunE: func(cmd *cobra.Command, args []string) error {
		spaceIds, err := parseSpaceInput(args)
		if err != nil {
			return err
		}

		if len(spaceIds) == 0 {
			return errors.New("at least one space id is required")
		}

		output := []*spinup.Resource{}
		for _, s := range spaceIds {
			log.Debugf("listing resources for space %s", s)

			resources, err := SpinupClient.Resources(s)
			if err != nil {
				return err
			}

			for _, r := range resources {
				if showFailedResources || r.Status != "failed" {
					output = append(output, r)
				}
			}
		}

		j, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return err
		}

		f := bufio.NewWriter(os.Stdout)
		defer f.Flush()
		f.Write(j)

		return nil
	},
}

var listImagesCmd = &cobra.Command{
	Use:   "images",
	Short: "List images in space",
	RunE: func(cmd *cobra.Command, args []string) error {
		spaceIds, err := parseSpaceInput(args)
		if err != nil {
			return err
		}

		if len(spaceIds) == 0 {
			return errors.New("space id is required")
		}

		log.Debugf("listing Images for space %s", spaceIds)

		type ImageOutput struct {
			*spinup.Image
			OfferingID   string `json:"offering_id"`
			OfferingName string `json:"offering_name"`
		}

		output := []*ImageOutput{}
		for _, s := range spaceIds {
			images := spinup.Images{}
			if err := SpinupClient.GetResource(map[string]string{"id": s}, &images); err != nil {
				return err
			}

			for _, i := range []*spinup.Image(images) {
				oID := i.Offering.ID.String()
				oName := i.Offering.Name
				i.Offering = nil
				output = append(output, &ImageOutput{i, oID, oName})
			}
		}

		j, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return err
		}

		f := bufio.NewWriter(os.Stdout)
		defer f.Flush()
		f.Write(j)

		return nil
	},
}
