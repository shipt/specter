'use strict';

/**
 * Config
 */
L.mapbox.accessToken = mpapi

const CHECK_VERSION_TIMEOUT_MS = 3000
const RECONNECT_TIMEOUT_MS = 3000
const STATUS_COLORS = {
  success: 'aqua',
  clientError: 'yellow',
  serverError: 'red'
};
const hqCircleLayer = new L.LayerGroup();
const circlesLayer = new L.LayerGroup();
let requestsPerSecCounter = 0;
let hqArray = [];
let circlesCanvas = L.canvas();
let ripplesEl;
let trafficSvg;

self.appVersion = ''
self.debug = false

/**
 * React (i.e. Presentation Layer)
 */
class App extends React.PureComponent {
  constructor(props) {
    super(props);
    this.state = {
      errorLogs: [],
      showErrorLogs: false,
      hasError: false,
      rpsCount: 0,
      appVersion: '',
      reconnectTimeoutMs: RECONNECT_TIMEOUT_MS
    }
  }

  componentDidMount() {
    // init #map
    this.map = L.mapbox.map('map', 'mapbox.dark', {
      center: [37.090240, -95.712891], // center on North America
      zoom: 6
    })
      .addLayer(hqCircleLayer)
      .addLayer(circlesLayer)

    trafficSvg = d3.select(this.map.getPanes().overlayPane).append('svg')
      .attr('class', 'leaflet-zoom-animated')
      .attr('width', window.innerWidth)
      .attr('height', window.innerHeight);

    ripplesEl = d3.select(this.map.getPanes().overlayPane).append('ripples')
      .attr('width', window.innerWidth)
      .attr('height', window.innerHeight)
      .call(drawOnCanvas);

    // read app version
    // MUST happen before HEAD /version, which polls every so often
    fetch('/version', { method: 'HEAD' })
      .then(res => {
         self.version = res.headers.get('Version')
      });

    this.setupWebSocket();
  }

  setupWebSocket() {
    try {
      let uri = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      uri += '//' + window.location.host;
      uri += window.location.pathname + 'ws';
      this.ws = new WebSocket(uri);

      this.ws.onmessage = (evt) => {
        // exit early if we get bad JSON data
        if (evt.data === '') return
        const wsData = JSON.parse(evt.data)
        const srcLatLng = new L.LatLng(wsData.SrcLat, wsData.SrcLong);
        const hqLatLng = new L.LatLng(wsData.DstLat, wsData.DstLong);
        requestsPerSecCounter++;
        // exit early if we get strange location data (i.e. from center of the earth to center of the earth)
        if (srcLatLng.lat === 0 && srcLatLng.lng === 0) {
          return
        } else if (srcLatLng.lat === 39.833 && srcLatLng.lng === -98.585) {
          return
        } else if (srcLatLng.lat === 37.751 && srcLatLng.lng === -97.822) {
          return
        }

        const srcPoint = this.map.latLngToLayerPoint(srcLatLng);
        const hqPoint = this.map.latLngToLayerPoint(hqLatLng);
        const httpStatus = Number(wsData.HTTPStatus);
        const midPoint = calcMidpoint(srcPoint, hqPoint);
        let statusColor = getStatusColor(httpStatus);

        if (shouldUpdateHQ(hqLatLng)) {
          addHQCircle(hqLatLng);
          hqArray.push(hqLatLng);
        }
        addCircleWithColor(srcLatLng, statusColor);
        handleRippleWithColor(srcPoint, statusColor);
        handleTrafficWithColor(srcPoint, midPoint, hqPoint, statusColor);

        // log server errors to UI
        if (httpStatus >= 500) {
          this.setErrorLog(`${httpStatus} Error from ${srcLatLng} to ${hqLatLng}`)
        }
      };

      // event setup for things like rps counters, etc
      this.ws.onopen = () => {
        this.setErrorLog(`WebSocket connected to ${uri}`);
        this.setState({ hasError: false, reconnectTimeoutMs: RECONNECT_TIMEOUT_MS });
        // setup counter
        this.requestCounter = setInterval(() => {
          this.setState(() => {
            const rpsCount = requestsPerSecCounter
            requestsPerSecCounter = 0
            return { rpsCount }
          })
        }, 1000)

        // this doubles as a kind of heartbeat to see if we still have connection to WS
        this.pollVersion = setInterval(() => fetchVersionService(this.ws), CHECK_VERSION_TIMEOUT_MS)
      };

      // clear/cleanup active events
      this.ws.onclose = () => {
        this.setErrorLog('on close callback invoked')
        this.setState({ hasError: true });
        this.retryWebSocket()
      };

      this.ws.onerror = () => {
        this.setErrorLog('on error callback invoked')
        this.setState({ hasError: true });
      };
    } catch (e) {
      this.setState({ hasError: true });
      this.setErrorLog(`caught error: ${e.message}`)
      this.retryWebSocket()
    }
  }

  retryWebSocket() {
    requestsPerSecCounter = 0;
    this.setState({ rpsCount: 0 });
    this.setErrorLog(`Websocket Terminated. Retrying in ${this.state.reconnectTimeoutMs}...`);
    clearInterval(this.requestCounter);
    clearInterval(this.pollVersion);
    // retry connection
    setTimeout(() => {
      this.setupWebSocket();
      // incremental back-off
      this.setState(prevState => ({ reconnectTimeoutMs: prevState.reconnectTimeoutMs + 1000 }))
    }, this.state.reconnectTimeoutMs);
  }

  toggleErrorLogger = () => {
    const showErrorLogs = !this.state.showErrorLogs
    if (showErrorLogs === false) this.setState({ errorLogs: [] })
    this.setState({ showErrorLogs });
    // also turns on browser console logs
    self.debug = showErrorLogs
  }

  setErrorLog = log => {
    if (this.state.showErrorLogs) {
      this.setState(prevState => ({ errorLogs: [...prevState.errorLogs, log] }));
      logger(log);
    }
  }

  render() {
    const { errorLogs, showErrorLogs, rpsCount, appVersion, hasError, reconnectTimeoutMs } = this.state
    return [
      React.createElement('div', { key: 'map', id: 'map' }),
      React.createElement(LegendComponent, { key: 'legend', onClick: this.toggleErrorLogger, rpsCount, isDebugging: showErrorLogs, appVersion }),
      React.createElement(ErrorLoggerComponent, { key: 'error-logger', logs: errorLogs, show: showErrorLogs }),
      React.createElement(ReconnectBannerComponent, { key: 'error-banner', show: hasError, reconnectTimeoutMs })
    ]
  }
}

const LegendComponent = React.memo(({ onClick, innerRef, rpsCount, isDebugging, appVersion }) => (
  <div id="legend" ref={innerRef}>
    <div style={{ float: 'right', color: '#888' }}>{appVersion}</div>
    <h2>HTTP Traffic Status</h2>
    <div>Data Center: <span style={{ color: 'white', fontWeight: 'bold' }}>white</span></div>
    <div>2xx Success: <span style={{ color: STATUS_COLORS.success, fontWeight: 'bold' }}>{STATUS_COLORS.success}</span></div>
    <div>4xx Client Error: <span style={{ color: STATUS_COLORS.clientError, fontWeight: 'bold' }}>{STATUS_COLORS.clientError}</span></div>
    <div>5xx Server Error: <span style={{ color: STATUS_COLORS.serverError, fontWeight: 'bold' }}>{STATUS_COLORS.serverError}</span></div>
    <div style={{ float: 'right' }}>Requests Per Second: {rpsCount}</div>
    <button onClick={onClick}>{isDebugging ? 'Disable' : 'Enable'} Debug Mode</button>
  </div>
))

const ErrorLoggerComponent = React.memo(({ logs, show }) => (
  <div id="error-logger" style={show ? { display: 'block' } : { display: 'none' }}>
    {logs.map((msg, i) => <div key={i}>{msg}</div>)}
  </div>
))

const ReconnectBannerComponent = React.memo(({ show, reconnectTimeoutMs }) => {
  if (!show) return null
  const [seconds, setSeconds] = React.useState(reconnectTimeoutMs / 1000)

  React.useEffect(() => {
    const countDown = setInterval(() => {
      setSeconds(prevSec => prevSec - 1)
    }, 1000)
    return function cleanup() {
      setSeconds(reconnectTimeoutMs / 1000)
      clearInterval(countDown)
    }
  }, [reconnectTimeoutMs])

  return (
    <div id="error-banner">
      {seconds < 0 ?
        'Unable to reconnect. Check your network connection.' :
        `Experiencing network issues. Attempting to reconnect in ${seconds}...`
      }
    </div>
  )
})

// render app!
ReactDOM.render(
  React.createElement(App),
  document.getElementById('app')
);

/**
 * Helper Functions
 * L.circle - https://leafletjs.com/reference-1.4.0.html#circle
 */
function addHQCircle(hqLatLng) {
  L.circle(hqLatLng, {
    renderer: circlesCanvas,
    radius: 50000,
    color: 'white',
    fillColor: 'white',
    fillOpacity: 0.8,
  }).addTo(hqCircleLayer);

  logger('added HQ location', hqLatLng);
}

function addCircleWithColor(srcLatLng, color) {
  const circleArray = circlesLayer.getLayers();

  // Only allow 10 circles to be on the map at a time
  if (circleArray.length >= 10) {
    circlesLayer.removeLayer(circleArray[0]);
  }

  L.circle(srcLatLng, {
    renderer: circlesCanvas,
    radius: 50000,
    color: color,
    fillColor: color,
    fillOpacity: 0.2,
  }).addTo(circlesLayer);
}

function handleRippleWithColor(srcPoint, color) {
  var x = srcPoint['x'];
  var y = srcPoint['y'];

  ripplesEl.append('ripple')
    .attr('x', x)
    .attr('y', y)
    .attr('r', 1e-6)
    .attr('stroke', color)
    .attr('opacity', 1)
    .transition()
    .duration(2000)
    .ease(Math.sqrt)
    .attr('r', 35)
    .attr('opacity', 1e-6)
    .remove();
}

function handleTrafficWithColor(srcPoint, midPoint, hqPoint, color) {
  var lineData = [srcPoint, midPoint, hqPoint]
  var lineFunction = d3.svg.line()
    .interpolate("basis")
    .x(function (d) { return d.x; })
    .y(function (d) { return d.y; });

  var lineGraph = trafficSvg.append('path')
    .attr('d', lineFunction(lineData))
    .attr('stroke-opacity', 0.8)
    .attr('stroke', color)
    .attr('stroke-width', 2)
    .attr('fill', 'none');

  var length = lineGraph.node().getTotalLength();

  lineGraph
    .attr('stroke-dasharray', length + ' ' + length)
    .attr('stroke-dashoffset', length)
    .transition()
    .duration(70)
    .ease('ease-in')
    .attr('stroke-dashoffset', 0)
    .each('end', function () {
      d3.select(this)
        .transition()
        .duration(100)
        .style('stroke-opacity', 0)
        .remove();
    });
}

function getStatusColor(httpStatus) {
  if (httpStatus >= 400 && httpStatus < 500) {
    return STATUS_COLORS.clientError
  } else if (httpStatus >= 500) {
    return STATUS_COLORS.serverError
  }
  return STATUS_COLORS.success;
}

function calcMidpoint(srcPoint, hqPoint) {
  var x1 = srcPoint['x'];
  var y1 = srcPoint['y'];
  var x2 = hqPoint['x'];
  var y2 = hqPoint['y'];
  var bendArray = [true, false];
  var bend = bendArray[Math.floor(Math.random() * bendArray.length)];

  if (y2 < y1 && x2 < x1) {
    var tmpy = y2;
    var tmpx = x2;
    x2 = x1;
    y2 = y1;
    x1 = tmpx;
    y1 = tmpy;
  }
  else if (y2 < y1) {
    var tmpy = y2;
    y2 = y1;
    y1 = tmpy;
  }
  else if (x2 < x1) {
    var tmpx = x2;
    x2 = x1;
    x1 = tmpx;
  }

  a = x2 + (x1 - x2)
  b = y2 + (y1 - y2)
  var radian = Math.atan(-((y2 - y1) / (x2 - x1)));
  var r = Math.sqrt(x2 - x1) + Math.sqrt(y2 - y1);
  var m1 = (x1 + x2) / 2;
  var m2 = (y1 + y2) / 2;

  var min = 2.5, max = 7.5;
  //var min = 1, max = 7;
  var arcIntensity = parseFloat((Math.random() * (max - min) + min).toFixed(2));

  if (bend === true) {
    var a = Math.floor(m1 - r * arcIntensity * Math.sin(radian));
    var b = Math.floor(m2 - r * arcIntensity * Math.cos(radian));
  } else {
    var a = Math.floor(m1 + r * arcIntensity * Math.sin(radian));
    var b = Math.floor(m2 + r * arcIntensity * Math.cos(radian));
  }

  return { "x": a, "y": b };
}

function shouldUpdateHQ(hqLatLng) {
  if (hqArray.some((loc) => hqLatLng.lat === loc.lat && hqLatLng.lng === loc.lng)) {
    // We already marked a dot for this hq so we dont need to again
    return false
  } else {
    return true
  }
}

function logger(msg1, msg2 = '') {
  if (self.debug) console.log(msg1, msg2)
}

function drawOnCanvas(selection) {
  selection.each(function () {
    var root = this,
      canvas = root.parentNode.appendChild(document.createElement('canvas')),
      ctx = canvas.getContext('2d');

    canvas.style.position = 'absolute';

    // It'd be nice to use DOM Mutation Events here instead.
    // However, they appear to arrive irregularly, causing choppy animation.
    d3.timer(redraw);

    // Clear the canvas and then iterate over child elements.
    function redraw() {
      canvas.width = root.getAttribute('width');
      canvas.height = root.getAttribute('height');
      for (var child = root.firstChild; child; child = child.nextSibling) draw(child);
    }

    // For now we only support a circle's ripple effect with strokeStyle.
    // But imagine extending this to arbitrary shapes and groups!
    function draw(element) {
      switch (element.tagName) {
        case 'RIPPLE': {
          ctx.globalAlpha = element.getAttribute('opacity')
          ctx.strokeStyle = element.getAttribute('stroke');
          ctx.beginPath();
          ctx.arc(element.getAttribute('x'), element.getAttribute('y'), element.getAttribute('r'), 0, 2 * Math.PI);
          ctx.stroke();
          break;
        }
      }
    }
  });
};

// // uses fetch with a timeout (if connection is lost)
// function fetchVersionService(webSocket) {
//   let didTimeOut = false;
//   new Promise((resolve, reject) => {
//     const timeout = setTimeout(() => {
//       didTimeOut = true;
//       reject(new Error('Request timed out'));
//     }, CHECK_VERSION_TIMEOUT_MS);

//     fetch('/version', { method: 'HEAD' })
//       .then(res => {
//         // Clear the timeout as cleanup
//         clearTimeout(timeout);
//         if (didTimeOut) return;
//         if (!res.ok) reject(new Error('Failed to connect'))
//         // Auto-Refresh on update!
//         // https://developer.mozilla.org/en-US/docs/Web/API/Location/reload
//         const version = res.headers.get('Version')
//         if (self.appVersion !== version) {
//           location.reload(true);
//         }
//         resolve(res);
//       })
//       .catch(err => {
//         // Rejection already happened with setTimeout
//         if (didTimeOut) return;
//         // Reject with error
//         reject(err);
//       });
//   })
//     .catch(err => {
//       // since this uses the same service as WS, if this is down, then we've lost connection
//       // so we should close the WS connection
//       webSocket.close();
//       logger(`fetchVersionService::${err.message}`)
//     });
// }
