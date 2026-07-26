(function (global) {
  'use strict';

  var STORAGE_DEVICE = 'member_device_id';
  var STORAGE_VISITOR = 'member_visitor_id';
  var STORAGE_SESSION = 'member_session_id';
  var HEARTBEAT_MS = 30000;

  function uuid() {
    if (global.crypto && crypto.randomUUID) return crypto.randomUUID();
    return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function (c) {
      var r = (Math.random() * 16) | 0;
      var v = c === 'x' ? r : (r & 0x3) | 0x8;
      return v.toString(16);
    });
  }

  function getScriptConfig() {
    var script = document.currentScript;
    if (!script) {
      var scripts = document.getElementsByTagName('script');
      script = scripts[scripts.length - 1];
    }
    return {
      siteKey: script && script.getAttribute('data-site-key'),
      endpoint: (script && script.getAttribute('data-endpoint')) || '/api/track',
    };
  }

  function storageGet(key) {
    try {
      var v = localStorage.getItem(key);
      if (v) return v;
    } catch (e) {}
    return null;
  }

  function storageSet(key, val) {
    try {
      localStorage.setItem(key, val);
    } catch (e) {}
  }

  function getDeviceId() {
    var id = storageGet(STORAGE_DEVICE);
    if (!id) {
      id = uuid();
      storageSet(STORAGE_DEVICE, id);
    }
    return id;
  }

  function getVisitorId() {
    var id = storageGet(STORAGE_VISITOR);
    if (!id) {
      id = uuid();
      storageSet(STORAGE_VISITOR, id);
    }
    return id;
  }

  function getSessionId() {
    var id = storageGet(STORAGE_SESSION);
    if (!id) {
      id = uuid();
      storageSet(STORAGE_SESSION, id);
    }
    return id;
  }

  function parseUtm() {
    var params = new URLSearchParams(global.location.search || '');
    return {
      utmSource: params.get('utm_source') || '',
      utmMedium: params.get('utm_medium') || '',
      utmCampaign: params.get('utm_campaign') || '',
    };
  }

  var cfg = getScriptConfig();
  var utm = parseUtm();
  var heartbeatTimer = null;
  var initialized = false;

  function send(eventType, props) {
    var payload = {
      eventType: eventType,
      pageUrl: global.location.href,
      referrer: document.referrer || '',
      utmSource: utm.utmSource,
      utmMedium: utm.utmMedium,
      utmCampaign: utm.utmCampaign,
      visitorId: getVisitorId(),
      sessionId: getSessionId(),
      deviceId: getDeviceId(),
      props: props || {},
    };
    if (cfg.siteKey) {
      payload.siteKey = cfg.siteKey;
    }
    payload.props.platform = navigator.platform || '';
    payload.props.userAgent = navigator.userAgent || '';
    return fetch(cfg.endpoint, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
      credentials: 'include',
    }).catch(function () {});
  }

  function trackPageView() {
    send('app.page_view', { path: global.location.pathname + global.location.hash });
  }

  function hookHistory() {
    var push = history.pushState;
    var replace = history.replaceState;
    history.pushState = function () {
      push.apply(history, arguments);
      trackPageView();
    };
    history.replaceState = function () {
      replace.apply(history, arguments);
      trackPageView();
    };
    global.addEventListener('popstate', trackPageView);
    global.addEventListener('hashchange', trackPageView);
  }

  function startSession() {
    send('session.start', {});
    if (heartbeatTimer) clearInterval(heartbeatTimer);
    heartbeatTimer = setInterval(function () {
      send('session.tick', {});
    }, HEARTBEAT_MS);
    global.addEventListener('beforeunload', function () {
      send('session.end', {});
    });
  }

  function init() {
    if (initialized) return;
    initialized = true;
    send('app.launch', {});
    startSession();
    hookHistory();
    trackPageView();
    global.addEventListener('error', function (ev) {
      send('app.crash', {
        message: ev.message,
        source: ev.filename,
        line: ev.lineno,
        col: ev.colno,
      });
    });
    global.addEventListener('unhandledrejection', function (ev) {
      send('app.crash', { message: String(ev.reason) });
    });
  }

  global.MemberTrack = function (name, props) {
    return send(name, props);
  };
  global.MemberAnalytics = { init: init, track: send, getDeviceId: getDeviceId };

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})(typeof window !== 'undefined' ? window : this);
