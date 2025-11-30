import { writable } from 'svelte/store';

export const isExtensionInstalled = writable(false);

let pollInterval: number | null = null;

/**
 * Check if extension is installed/connected
 * Simple heartbeat check - no events, no complexity
 */
export async function checkExtension() {
  try {
    const installed = await window.go.main.App.CheckChromeExtension();
    isExtensionInstalled.set(installed);

    if (!installed) {
      // Not connected, start polling if not already
      startPolling();
    } else {
      // Connected! But keep polling to detect disconnects
      startPolling();
    }
  } catch (error) {
    console.error('Error checking extension:', error);
    isExtensionInstalled.set(false);
    startPolling();
  }
}

/**
 * Start polling for extension status
 * Checks every 3 seconds
 */
function startPolling() {
  if (pollInterval !== null) {
    return; // Already polling
  }

  pollInterval = setInterval(async () => {
    try {
      const installed = await window.go.main.App.CheckChromeExtension();
      isExtensionInstalled.set(installed);
    } catch (error) {
      console.error('Polling error:', error);
      isExtensionInstalled.set(false);
    }
  }, 3000); // Check every 3 seconds
}

/**
 * Stop polling (cleanup)
 */
export function stopPolling() {
  if (pollInterval !== null) {
    clearInterval(pollInterval);
    pollInterval = null;
  }
}

/**
 * Register the extension with backend
 * Creates native-host.json manifest
 */
export async function registerExtension() {
  try {
    await window.go.main.App.RegisterExtension(
      'hkanepohpflociaodcicmmfbdaohpceo'
    );
    await window.go.main.App.RegisterExtension(
      'gpaafgcbiejjpfdgmjglehboafdicdjb'
    );
    console.log('Extension registered');
    // Force immediate re-check
    await checkExtension();
  } catch (error) {
    console.error('Error registering extension:', error);
  }
}
