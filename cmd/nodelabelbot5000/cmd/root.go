package cmd

import (
	"sync"

	"github.com/samsung-cnct/nodelabelbot5000/pkg/util"
	"github.com/samsung-cnct/nodelabelbot5000/pkg/util/k8sutil"
	"github.com/spf13/viper"

	"flag"
	"fmt"
	"github.com/juju/loggo"
	"github.com/samsung-cnct/nodelabelbot5000/pkg/controllers/nodelabelbot5000"
	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/rest"
	"os"
	"strings"
)

var (
	logger loggo.Logger
	config *rest.Config

	rootCmd = &cobra.Command{
		Use:   "operator",
		Short: "CMA Operator",
		Long:  `The CMA Operator`,
		Run: func(cmd *cobra.Command, args []string) {
			operator()
		},
	}
)

func init() {
	viper.SetEnvPrefix("cmaoperator")
	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)

	rootCmd.Flags().String("kubeconfig", "", "Location of kubeconfig file")
	rootCmd.Flags().String("kubernetes-namespace", "default", "Namespace to operate on")

	viper.BindPFlag("kubeconfig", rootCmd.Flags().Lookup("kubeconfig"))
	viper.BindPFlag("kubernetes-namespace", rootCmd.Flags().Lookup("kubernetes-namespace"))

	viper.AutomaticEnv()
	rootCmd.Flags().AddGoFlagSet(flag.CommandLine)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func operator() {
	var err error
	logger := util.GetModuleLogger("cmd.nodelabelbot5000", loggo.INFO)
	//viperInit()

	// get flags
	portNumber := viper.GetInt("port")
	kubeconfigLocation := viper.GetString("kubeconfig")

	// Debug for now
	logger.Infof("Parsed Variables: \n  Port: %d \n  Kubeconfig: %s", portNumber, kubeconfigLocation)

	k8sutil.KubeConfigLocation = kubeconfigLocation
	k8sutil.DefaultConfig, err = k8sutil.GenerateKubernetesConfig()

	if err != nil {
		logger.Infof("Was unable to generate a valid kubernetes default config, some functionality may be broken.  Error was %v", err)
	}

	var wg sync.WaitGroup
	stop := make(chan struct{})

	logger.Infof("Starting the NodeLabelBot5000 Controller")
	nodelabelbot5000Controller, err := nodelabelbot5000.NewNodeLabelBot5000Controller(nil)
	if err != nil {
		logger.Criticalf("Could not create a nodelabelbot5000")
		os.Exit(-1)
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		nodelabelbot5000Controller.Run(3, stop)
	}()

	<-stop
	logger.Infof("Wating for controllers to shut down gracefully")
	wg.Wait()
}
