package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/redhat-developer/odo/pkg/application"
	"github.com/redhat-developer/odo/pkg/project"
	"github.com/redhat-developer/odo/pkg/url"
	"github.com/spf13/cobra"
)

var (
	urlComponent       string
	urlApplication     string
	urlForceDeleteFlag bool
)

var urlCmd = &cobra.Command{
	Use:   "url",
	Short: "Expose component to the outside world",
	Long: `Expose component to the outside world.

The URLs that are generated using this command, can be used to access the deployed components from outside the cluster.`,
	Example: fmt.Sprintf("%s\n%s\n%s",
		urlCreateCmd.Example,
		urlDeleteCmd.Example,
		urlListCmd.Example),
}

var urlCreateCmd = &cobra.Command{
	Use:   "create [component name]",
	Short: "Create a URL for a component",
	Long: `Create a URL for a component.

The created URL can be used to access the specified component from outside the OpenShift cluster.
`,
	Example: `  # Create a URL for the current component.
  odo url create

  # Create a URL with a specific name
  odo url create example

  # Create a URL with a specific name for component frontend
  odo url create example --component frontend
	`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := getOcClient()
		applicationName, err := application.GetCurrent(client)
		checkError(err, "")
		projectName := project.GetCurrent(client)

		var app string
		var urlName string

		if urlApplication == "" {
			app, err = application.GetCurrent(client)
			checkError(err, "")
		} else {
			exists, err := application.Exists(client, urlApplication)
			checkError(err, "unable to check if the application exists or not")
			if !exists {
				fmt.Printf("The application %s does not exists in the project %s\n", urlApplication, projectName)
				os.Exit(1)
			}
			app = urlApplication
		}

		componentName := getComponent(client, urlComponent, app, projectName)

		switch len(args) {
		case 0:
			urlName = componentName
		case 1:
			urlName = args[0]
		default:
			fmt.Println("unable to get component")
			os.Exit(1)
		}

		exists, err := url.Exists(client, urlName, "", applicationName)

		if exists {
			fmt.Printf("The url %s already exists in the application: %s\n", urlName, applicationName)
			os.Exit(1)
		}

		fmt.Printf("Adding URL to component: %v\n", componentName)
		urlRoute, err := url.Create(client, urlName, componentName, applicationName)
		checkError(err, "")
		fmt.Printf("URL created for component: %v\n\n"+
			"%v - %v\n", componentName, urlRoute.Name, url.GetUrlString(*urlRoute))
	},
}

var urlDeleteCmd = &cobra.Command{
	Use:   "delete <url-name>",
	Short: "Delete a URL",
	Long:  `Delete the given URL, hence making the service inaccessible.`,
	Example: `  # Delete a URL to a component
  odo url delete myurl
	`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		// Initialization
		client := getOcClient()
		projectName := project.GetCurrent(client)
		applicationName, err := application.GetCurrent(client)
		checkError(err, "")

		componentName := getComponent(client, urlComponent, applicationName, projectName)

		urlName := args[0]

		exists, err := url.Exists(client, urlName, componentName, applicationName)
		checkError(err, "")

		if !exists {
			fmt.Printf("The URL %s does not exist within the component %s\n", urlName, componentName)
			os.Exit(1)
		}

		var confirmDeletion string
		if urlForceDeleteFlag {
			confirmDeletion = "y"
		} else {
			fmt.Printf("Are you sure you want to delete the url %v? [y/N] ", urlName)
			fmt.Scanln(&confirmDeletion)
		}

		if strings.ToLower(confirmDeletion) == "y" {

			err = url.Delete(client, urlName, applicationName)
			checkError(err, "")
			fmt.Printf("Deleted URL: %v\n", urlName)
		} else {
			fmt.Printf("Aborting deletion of url: %v\n", urlName)
		}
	},
}

var urlListCmd = &cobra.Command{
	Use:   "list",
	Short: "List URLs",
	Long:  `Lists all the available URLs which can be used to access the components.`,
	Example: ` # List the available URLs
  odo url list
	`,
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		client := getOcClient()
		projectName := project.GetCurrent(client)

		var app string
		var err error

		if urlApplication == "" {
			app, err = application.GetCurrent(client)
			checkError(err, "")
		} else {
			exists, err := application.Exists(client, urlApplication)
			checkError(err, "unable to check if the application exists or not")
			if !exists {
				fmt.Printf("The application %s does not exists in the project %s\n", urlApplication, projectName)
				os.Exit(1)
			}
			app = urlApplication
		}

		componentName := getComponent(client, urlComponent, app, projectName)

		urls, err := url.List(client, componentName, app)
		checkError(err, "")

		if len(urls) == 0 {
			fmt.Printf("No URLs found for component %v in application %v\n", componentName, app)
		} else {
			fmt.Printf("Found the following URLs for component %v in application %v:\n", componentName, app)
			for _, u := range urls {
				fmt.Printf("%v - %v\n", u.Name, url.GetUrlString(u))
			}
		}
	},
}

func init() {
	urlCreateCmd.Flags().StringVarP(&urlApplication, "application", "a", "", "create url for application")
	urlCreateCmd.Flags().StringVarP(&urlComponent, "component", "c", "", "create url for component")

	urlDeleteCmd.Flags().BoolVarP(&urlForceDeleteFlag, "force", "f", false, "Delete url without prompting")
	urlDeleteCmd.Flags().StringVarP(&urlComponent, "component", "c", "", "delete url for component")

	urlListCmd.Flags().StringVarP(&urlApplication, "application", "a", "", "list URLs for application")
	urlListCmd.Flags().StringVarP(&urlComponent, "component", "c", "", "list URLs for component")

	urlCmd.AddCommand(urlListCmd)
	urlCmd.AddCommand(urlDeleteCmd)
	urlCmd.AddCommand(urlCreateCmd)

	// Add a defined annotation in order to appear in the help menu
	urlCmd.Annotations = map[string]string{"command": "other"}
	urlCmd.SetUsageTemplate(cmdUsageTemplate)

	rootCmd.AddCommand(urlCmd)
}
