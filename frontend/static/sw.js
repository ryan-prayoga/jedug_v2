self.addEventListener("push", (event) => {
  event.waitUntil(handlePush(event));
});

self.addEventListener("notificationclick", (event) => {
  event.notification.close();
  event.waitUntil(handleNotificationClick(event));
});

async function handlePush(event) {
  const payload = parsePayload(event);
  if (!payload) {
    return;
  }

  const clientsList = await self.clients.matchAll({
    type: "window",
    includeUncontrolled: true,
  });

  const visibleClients = clientsList.filter(
    (client) => client.visibilityState === "visible",
  );

  if (visibleClients.length > 0) {
    for (const client of visibleClients) {
      client.postMessage({
        type: "jedug:push-received",
        payload,
      });
    }
    return;
  }

  await self.registration.showNotification(payload.title, {
    body: payload.body,
    icon: "/push-icon.svg",
    tag: `issue-${payload.issue_id}-${payload.type}`,
    data: payload,
  });
}

async function handleNotificationClick(event) {
  const payload = event.notification?.data;
  const targetURL = payload?.url || "/";
  const clientsList = await self.clients.matchAll({
    type: "window",
    includeUncontrolled: true,
  });

  const exactClient = clientsList.find((client) => sameURL(client.url, targetURL));
  if (exactClient) {
    await exactClient.focus();
    exactClient.postMessage({
      type: "jedug:push-open-issue",
      issue_id: payload?.issue_id,
      url: targetURL,
    });
    return;
  }

  const sameOriginClient = clientsList.find((client) => sameOrigin(client.url, targetURL));
  if (sameOriginClient) {
    await sameOriginClient.focus();
    if ("navigate" in sameOriginClient) {
      await sameOriginClient.navigate(targetURL);
    }
    sameOriginClient.postMessage({
      type: "jedug:push-open-issue",
      issue_id: payload?.issue_id,
      url: targetURL,
    });
    return;
  }

  await self.clients.openWindow(targetURL);
}

function parsePayload(event) {
  if (!event.data) return null;

  try {
    const payload = event.data.json();
    if (!payload?.title || !payload?.url || !payload?.issue_id) {
      return null;
    }
    return payload;
  } catch {
    return null;
  }
}

function sameURL(currentURL, targetURL) {
  return new URL(currentURL).href === new URL(targetURL, self.location.origin).href;
}

function sameOrigin(currentURL, targetURL) {
  return new URL(currentURL).origin === new URL(targetURL, self.location.origin).origin;
}
