const hostName = 'com.infraflakes.Veda';
let port;
let webBlocklist = [];
let isPortConnected = false;

function connect() {
  try {
    port = chrome.runtime.connectNative(hostName);
    isPortConnected = true;

    let isExpectedDisconnect = false;

    port.onMessage.addListener((msg) => {
      if (msg.type === 'web_blocklist') {
        webBlocklist = msg.payload || [];
      } else if (msg.type === 'stopping') {
        // App is stopping intentionally, don't reconnect
        isExpectedDisconnect = true;
      }
    });

    port.onDisconnect.addListener(() => {
      isPortConnected = false;
      if (chrome.runtime.lastError) {
      }

      // Only reconnect if it wasn't an expected disconnect
      if (!isExpectedDisconnect) {
        // Retry native messaging connection after 5 seconds
        // This reconnection interval allows the extension to resume functionality quickly
        // if Veda daemon restarts or connection is lost
        setTimeout(connect, 5000);
      }
    });

    // Request the blocklist on connection.
    port.postMessage({ type: 'get_web_blocklist' });
  } catch (err) {
    isPortConnected = false;
    // Retry connection if initial connect fails (e.g. daemon not running yet)
    setTimeout(connect, 5000);
  }
}

connect();

// Listen for messages from the web GUI for installation detection.
chrome.runtime.onMessageExternal.addListener((request, sender, sendResponse) => {
  if (request.message === 'is_installed') {
    sendResponse({
      status: 'installed',
      version: chrome.runtime.getManifest().version,
      id: chrome.runtime.id,
    });
  }
  // Return true to indicate you wish to send a response asynchronously
  return true;
});

// Listen for messages from the popup.
chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
  if (request.type === 'add_to_web_blocklist') {
    if (port && isPortConnected) {
      port.postMessage(request);
      // Also add it to the local blocklist immediately for a faster response.
      if (!webBlocklist.includes(request.payload)) {
        webBlocklist.push(request.payload);
      }
      sendResponse({ status: 'ok' });
    } else {
      sendResponse({ status: 'error', message: 'Port not connected' });
    }
  } else if (request.type === 'log_web_metadata') {
    if (port && isPortConnected) {
      port.postMessage(request);
    }
  }
});

chrome.tabs.onUpdated.addListener((tabId, changeInfo, tab) => {
  // Inject extension ID into the web UI
  if (
    changeInfo.status === 'complete' &&
    tab.url &&
    tab.url.startsWith('http://127.0.0.1:58141')
  ) {
    chrome.scripting.executeScript({
      target: { tabId: tabId },
      func: () => {
        const extensionId = chrome.runtime.id;
        // Remove existing div if it's there
        const existingDiv = document.getElementById('Veda-extension-id');
        if (existingDiv) {
          existingDiv.remove();
        }
        const idDiv = document.createElement('div');
        idDiv.id = 'Veda-extension-id';
        idDiv.textContent = extensionId;
        idDiv.style.display = 'none';
        document.body.appendChild(idDiv);
      },
    });
  }

  const blockedPage = chrome.runtime.getURL('blocked.html');
  if (tab.url === blockedPage) {
    return;
  }

  // Check if the site is in the blocklist.
  if (tab.url) {
    try {
      const url = new URL(tab.url);
      const domain = url.hostname;
      if (webBlocklist.includes(domain)) {
        chrome.tabs.update(tabId, { url: blockedPage });
        return;
      }
    } catch (e) {
      // Ignore invalid URLs
    }
  }

  // Only log when the tab is fully loaded and has a valid URL.
  if (changeInfo.status === 'complete' && tab.url && (tab.url.startsWith('http') || tab.url.startsWith('https'))) {
    if (port && isPortConnected) {
      port.postMessage({
        type: 'log_url',
        payload: {
          url: tab.url,
          title: tab.title,
          visitTime: Math.floor(Date.now() / 1000)
        }
      });

      // Inject a script to get the title and favicon
      chrome.scripting.executeScript({
        target: { tabId: tabId },
        func: () => {
          function getFaviconUrl() {
            let faviconUrl = '';
            const linkElements = document.head.querySelectorAll('link[rel*="icon"]');
            for (const link of linkElements) {
              if (link.href) {
                faviconUrl = link.href;
                break;
              }
            }
            if (!faviconUrl) {
              faviconUrl = `${window.location.origin}/favicon.ico`;
            }
            return faviconUrl;
          }

          chrome.runtime.sendMessage({
            type: 'log_web_metadata',
            payload: {
              domain: window.location.hostname,
              title: document.title,
              iconUrl: getFaviconUrl()
            }
          });
        }
      });
    }
  }
});
