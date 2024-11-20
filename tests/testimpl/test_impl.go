package testimpl

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/launchbynttdata/lcaf-component-terratest/types"
	"github.com/stretchr/testify/assert"
)

func TestFunctionApp(t *testing.T, ctx types.TestContext) {
	ctx.EnabledOnlyForTests(t, "complete", "windows_func_app")

	subscriptionId := os.Getenv("ARM_SUBSCRIPTION_ID")
	if len(subscriptionId) == 0 {
		t.Fatal("ARM_SUBSCRIPTION_ID environment variable is not set")
	}

	webAppHostname := terraform.Output(t, ctx.TerratestTerraformOptions(), "web_app_slot_default_hostname")

	status := retry.DoWithRetry(t, "Check if the function app is up and running", 6, 10*time.Second, func() (string, error) {
		res, err := http.Get(fmt.Sprintf("https://%s", webAppHostname))
		return strconv.FormatInt(int64(res.StatusCode), 10), err
	})

	assert.Equal(t, "200", status)
}

func TestComposablePrivateFuncApp(t *testing.T, ctx types.TestContext) {
	ctx.EnabledOnlyForTests(t, "private_func_app")

	subscriptionId := os.Getenv("ARM_SUBSCRIPTION_ID")
	if len(subscriptionId) == 0 {
		t.Fatal("ARM_SUBSCRIPTION_ID environment variable is not set")
	}

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		t.Fatalf("Failed to get Azure credentials: %v", err)
	}

	webAppClient, err := armappservice.NewWebAppsClient(subscriptionId, cred, nil)
	if err != nil {
		t.Fatalf("Failed to create Azure Function App client: %v", err)
	}

	t.Run("TestDefaultHostName", func(t *testing.T) {
		defaultHostname := terraform.Output(t, ctx.TerratestTerraformOptions(), "web_app_slot_default_hostname")
		resourceGroupName := terraform.Output(t, ctx.TerratestTerraformOptions(), "resource_group_name")
		slotName := terraform.Output(t, ctx.TerratestTerraformOptions(), "web_app_slot_name")

		azureWebApp, err := webAppClient.Get(context.Background(), resourceGroupName, slotName, nil)
		if err != nil {
			t.Fatalf("Failed to get Azure Function App: %v", err)
		}
		assert.Equal(t, *azureWebApp.Properties.DefaultHostName, defaultHostname, "Expected Default hostname did not match actual Default hostname!")
	})

	t.Run("TestFunctionAppID", func(t *testing.T) {

		resourceGroupName := terraform.Output(t, ctx.TerratestTerraformOptions(), "resource_group_name")
		slotName := terraform.Output(t, ctx.TerratestTerraformOptions(), "web_app_slot_name")
		functionAppID := terraform.Output(t, ctx.TerratestTerraformOptions(), "web_app_slot_id")

		azureWebApp, err := webAppClient.Get(context.Background(), resourceGroupName, slotName, nil)
		if err != nil {
			t.Fatalf("Failed to get Azure Function App: %v", err)
		}
		assert.Equal(t, *azureWebApp.ID, functionAppID, "Expected ID did not match actual ID!")
	})

	t.Run("TestSlotUrl", func(t *testing.T) {
		slotUrl := terraform.Output(t, ctx.TerratestTerraformOptions(), "web_app_slot_default_hostname")
		resourceGroupName := terraform.Output(t, ctx.TerratestTerraformOptions(), "resource_group_name")
		slotName := terraform.Output(t, ctx.TerratestTerraformOptions(), "web_app_slot_name")

		azureWebApp, err := webAppClient.Get(context.Background(), resourceGroupName, slotName, nil)
		if err != nil {
			t.Fatalf("Failed to get Azure Function App: %v", err)
		}
		assert.Equal(t, *azureWebApp.Properties.HostNameSSLStates[0].Name, slotUrl, "Expected URL did not match actual URL!")
	})

}
