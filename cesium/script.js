// Your access token can be found at: https://cesium.com/ion/tokens.
Cesium.Ion.defaultAccessToken = '{{.Token}}';
// Initialize the Cesium Viewer in the HTML element with the `cesiumContainer` ID.
const viewer = new Cesium.Viewer('cesiumContainer', {
    terrainProvider: Cesium.createWorldTerrain()
});
// Add Cesium OSM Buildings, a global 3D buildings layer.
const buildingTileset = viewer.scene.primitives.add(Cesium.createOsmBuildings());
viewer.camera.flyTo({
    destination: Cesium.Cartesian3.fromDegrees({{.Camera.Long}}, {{.Camera.Lat}}, {{.Camera.Alt}}),
    orientation: {
        heading: Cesium.Math.toRadians({{.Camera.Heading}}),
        pitch: Cesium.Math.toRadians({{.Camera.Pitch}}),
    }
});

const positionProperty = new Cesium.SampledPositionProperty();

function main() {
    if (!("WebSocket" in window)) {
        alert("WebSocket NOT supported by your Browser!");
        return;
    }

    var ws = new WebSocket(window.location.origin.replace("http", "ws") + "/ws");

    ws.onopen = function () {
        console.log(`ws connected`);
    };

    ws.onmessage = function (evt) {
        const msg = JSON.parse(evt.data);
        const position = Cesium.Cartesian3.fromDegrees(msg.Long, msg.Lat, msg.Alt);
        const start = Cesium.JulianDate.fromIso8601(msg.Time);
        const stop = Cesium.JulianDate.addSeconds(start, {{.LiveTime}}, new Cesium.JulianDate());
        positionProperty.addSample(start, position);
        const id = msg.ID;

        const entity = viewer.entities.getOrCreateEntity(id);
        entity.description = `${msg.ID}`;
        entity.position = position;
        entity.path = new Cesium.PathGraphics({ width: 3 });

        // Make the items disappear if they are not available for {{.LiveTime}} seconds.
        entity.availability = new Cesium.TimeIntervalCollection([ new Cesium.TimeInterval({ start: start, stop: stop }) ]);
        
        entity.label = {
            text: `${msg.ID}\n` + 
                  `Alt: ${msg.Alt}m\n` +
                  `Speed: ${msg.GroundSpeed}ms/s\n` +
                  `Vario: ${msg.Climb}m/s`,
            font: '20pt monospace',
            pixelOffset: new Cesium.Cartesian2(0, -50),
            scaleByDistance: new Cesium.NearFarScalar(0.0, 1.0, 1.0e4, 0.2)
        };
        if (msg.Type == "glider") {
            entity.model = { uri: "/models/glider/scene.gltf"};
        } else if (msg.Type == "towplane") {
            entity.model = { uri: "/models/towplane/towplane.glb"};
        } else {
            entity.point = { pixelSize: 10, color: Cesium.Color.RED };
        }
    };

    ws.onclose = function () {
        console.log(`ws disconnected`);
    };
}
main();