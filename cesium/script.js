// Consts from template.
const cameraLong = {{.Camera.Long }};
const cameraLat = {{.Camera.Lat }};
const cameraAlt = {{.Camera.Alt }};
const cameraHeading = {{.Camera.Heading }};
const cameraPitch = {{.Camera.Pitch }};

const liveTime = {{.LiveTime }};
const altFix = {{.AltFix }};


// Your access token can be found at: https://cesium.com/ion/tokens.
Cesium.Ion.defaultAccessToken = '{{.Token}}';
// Initialize the Cesium Viewer in the HTML element with the `cesiumContainer` ID.
const viewer = new Cesium.Viewer('cesiumContainer', {
    terrainProvider: Cesium.createWorldTerrain()
});
// Add Cesium OSM Buildings, a global 3D buildings layer.
const buildingTileset = viewer.scene.primitives.add(Cesium.createOsmBuildings());
viewer.camera.flyTo({
    destination: Cesium.Cartesian3.fromDegrees(cameraLong, cameraLat, cameraAlt),
    orientation: {
        heading: Cesium.Math.toRadians(cameraHeading),
        pitch: Cesium.Math.toRadians(cameraPitch),
    }
});

// Set time slider
const start = Cesium.JulianDate.now(new Cesium.JulianDate());
const stop = Cesium.JulianDate.addSeconds(start, 60 * 60, new Cesium.JulianDate());
viewer.clock.startTime = start.clone();
viewer.clock.stopTime = stop.clone();
viewer.clock.currentTime = start.clone();
viewer.timeline.zoomTo(start.clone(), stop.clone());
viewer.clock.shouldAnimate = true; // Start playing the scene.

const positionProperty = new Cesium.SampledPositionProperty();

function model(airplaneType) {
    switch (airplaneType) {
        case "glider":
            return new Cesium.ModelGraphics({
                uri: "/models/glider/scene.gltf",
                scale: 20,
            });
        case "towplane":
            return new Cesium.ModelGraphics({
                uri: "/models/towplane/towplane.gltf",
                allowPicking: 1,
                scale: 15,
            });
        default:
            return new Cesium.ModelGraphics({
                uri: "/models/ufo/scene.gltf",
            });
    }
}

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
        const position = Cesium.Cartesian3.fromDegrees(msg.Long, msg.Lat, msg.Alt + altFix);
        const start = Cesium.JulianDate.fromIso8601(msg.Time);
        const stop = Cesium.JulianDate.addSeconds(start, liveTime, new Cesium.JulianDate());
        positionProperty.addSample(start, position);
        const id = msg.ID;

        const entity = viewer.entities.getOrCreateEntity(id);
        entity.description = `${msg.ID}`;
        entity.position = position;
        entity.path = new Cesium.PathGraphics({ width: 3 });

        // Make the items disappear if they are not available for {{.LiveTime}} seconds.
        entity.availability = new Cesium.TimeIntervalCollection([new Cesium.TimeInterval({ start: start, stop: stop })]);

        entity.label = {
            text: `${msg.ID}\n` +
                `Alt: ${msg.Alt}m\n` +
                `Speed: ${msg.GroundSpeed}ms/s\n` +
                `Vario: ${msg.Climb}m/s`,
            font: '20pt monospace',
            pixelOffset: new Cesium.Cartesian2(0, -50),
            scaleByDistance: new Cesium.NearFarScalar(0.0, 1.0, 1.0e4, 0.2)
        };
        entity.model = model(msg.Type)
    };

    ws.onclose = function () {
        console.log(`ws disconnected`);
    };
}
main();

// Allowed flying areas
viewer.entities.add({
    position: Cesium.Cartesian3.fromDegrees(35.22903, 32.59705, 70.0),
    ellipsoid: {
        radii: new Cesium.Cartesian3(3200.0, 3200.0, 10000.0),
        minimumClock: Cesium.Math.toRadians(0.0),
        maximumClock: Cesium.Math.toRadians(-180.0),
        minimumCone: Cesium.Math.toRadians(90.0),
        maximumCone: Cesium.Math.toRadians(85.6),
        material: Cesium.Color.BLACK.withAlpha(0.3),
    },
});

viewer.entities.add({
    position: Cesium.Cartesian3.fromDegrees(35.22903, 32.59705, 70.0),
    ellipsoid: {
        radii: new Cesium.Cartesian3(2000.0, 2000.0, 10000.0),
        minimumClock: Cesium.Math.toRadians(0.0),
        maximumClock: Cesium.Math.toRadians(180.0),
        minimumCone: Cesium.Math.toRadians(90.0),
        maximumCone: Cesium.Math.toRadians(85.6),
        material: Cesium.Color.BLACK.withAlpha(0.3),
    },
});

viewer.entities.add({
    position: Cesium.Cartesian3.fromDegrees(35.22903, 32.59705, 451.0),
    plane: {
        plane: new Cesium.Plane(Cesium.Cartesian3.UNIT_Y, 0.0),
        dimensions: new Cesium.Cartesian2(6400.0, 762.0),
        material: Cesium.Color.BLACK.withAlpha(0.3),
    },
});

viewer.entities.add({
    polygon: {
        hierarchy: {
            positions: Cesium.Cartesian3.fromDegreesArray([
                //Lower left corner
                35.20000,
                32.58500,
                //Upper left corner
                35.20000,
                32.60000,
                //Upper right corner
                35.25000,
                32.60000,
                //Lower right corner
                35.25000,
                32.58500,
            ]),
            holes: [
                {
                    positions: Cesium.Cartesian3.fromDegreesArray([
                        //Lower left corner
                        35.20500,
                        32.58900,
                        //Upper left corner
                        35.20500,
                        32.59500,
                        //Upper right corner
                        35.24500,
                        32.59500,
                        //Lower right corner
                        35.24500,
                        32.58900,
                    ]),
                },
            ],
        },
        material: Cesium.Color.BLUE.withAlpha(0.05),
        height: 396.24,
        extrudedHeight: 346.24,
        outline: true, // height is required for outline to display
    },
});