// Consts from template.
const cameraLong = {{.Camera.Long }};
const cameraLat = {{.Camera.Lat }};
const cameraAlt = {{.Camera.Alt }};
const cameraHeading = {{.Camera.Heading }};
const cameraPitch = {{.Camera.Pitch }};

const altFix = {{.AltFix }};
const pathLength = {{.PathLength }};
const minGroundSpeed = {{.MinGroundSpeed }};


// Your access token can be found at: https://cesium.com/ion/tokens.
Cesium.Ion.defaultAccessToken = '{{.Token}}';
// Initialize the Cesium Viewer in the HTML element with the `cesiumContainer` ID.
const viewer = new Cesium.Viewer('cesiumContainer', {
    terrainProvider: Cesium.createWorldTerrain()
});
// Add Cesium OSM Buildings, a global 3D buildings layer.
viewer.scene.primitives.add(Cesium.createOsmBuildings());
viewer.camera.flyTo({
    destination: Cesium.Cartesian3.fromDegrees(cameraLong, cameraLat, cameraAlt),
    orientation: {
        heading: Cesium.Math.toRadians(cameraHeading),
        pitch: Cesium.Math.toRadians(cameraPitch),
    }
});

// Set time slider
const start = Cesium.JulianDate.now(new Cesium.JulianDate());
const stop = Cesium.JulianDate.addSeconds(start, 3 * 60 * 60, new Cesium.JulianDate());
viewer.clock.startTime = start.clone();
viewer.clock.stopTime = stop.clone();
viewer.clock.currentTime = start.clone();
viewer.clock.shouldAnimate = true; // Start playing the scene.
viewer.timeline.zoomTo(start.clone(), stop.clone());

function model(airplaneType) {
    switch (airplaneType) {
        case "glider":
            return new Cesium.ModelGraphics({
                uri: "/models/glider/scene.gltf",
                allowPicking: 1,
                minimumPixelSize: 10,
            });
        case "towplane":
            return new Cesium.ModelGraphics({
                uri: "/models/towplane/towplane.gltf",
                allowPicking: 1,
                minimumPixelSize: 25,
            });
        default:
            return new Cesium.ModelGraphics({
                uri: "/models/ufo/scene.gltf",
                allowPicking: 1,
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
        const time = Cesium.JulianDate.fromIso8601(msg.Time);
        const id = msg.ID;

        if (msg.GroundSpeed < minGroundSpeed) {
            return;
        }

        if (!viewer.entities.getById(id)) {
            console.log(`Creating ${id}.`);
            var posProp = new Cesium.SampledPositionProperty();
            viewer.entities.add({
                id: id,
                position: posProp,
                model: model(msg.Type),
                description: `${id}`,
                orientation: new Cesium.VelocityOrientationProperty(posProp),
                path: new Cesium.PathGraphics({
                    width: 2,
                    trailTime: pathLength,
                    material: new Cesium.ColorMaterialProperty(nextColor()),
                }),
                availability: new Cesium.TimeIntervalCollection([new Cesium.TimeInterval({
                    start: start.clone(),
                    stop: stop.clone(),
                })]),
            });
        }

        const entity = viewer.entities.getById(id);
        entity.position.addSample(time, position);
        entity.label = {
            text: `${msg.ID}\n` +
                `Alt: ${msg.Alt}m\n` +
                `Speed: ${msg.GroundSpeed}ms/s\n` +
                `Vario: ${msg.Climb}m/s\n` +
                `Heading: ${msg.Dir}°`,
            font: '20pt monospace',
            pixelOffset: new Cesium.Cartesian2(0, -50),
            scaleByDistance: new Cesium.NearFarScalar(0.0, 1.0, 1.0e4, 0.2)
        };
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
    },
});

const colors = [
    Cesium.Color.ALICEBLUE,
    Cesium.Color.ANTIQUEWHITE,
    Cesium.Color.AQUA,
    Cesium.Color.AQUAMARINE,
    Cesium.Color.AZURE,
    Cesium.Color.BEIGE,
    Cesium.Color.BISQUE,
    Cesium.Color.BLACK,
    Cesium.Color.BLANCHEDALMOND,
    Cesium.Color.BLUE,
    Cesium.Color.BLUEVIOLET,
    Cesium.Color.BROWN,
    Cesium.Color.BURLYWOOD,
    Cesium.Color.CADETBLUE,
    Cesium.Color.CHARTREUSE,
    Cesium.Color.CHOCOLATE,
    Cesium.Color.CORAL,
    Cesium.Color.CORNFLOWERBLUE,
    Cesium.Color.CORNSILK,
    Cesium.Color.CRIMSON,
    Cesium.Color.CYAN,
    Cesium.Color.DARKBLUE,
    Cesium.Color.DARKCYAN,
    Cesium.Color.DARKGOLDENROD,
    Cesium.Color.DARKGRAY,
    Cesium.Color.DARKGREEN,
    Cesium.Color.DARKKHAKI,
    Cesium.Color.DARKMAGENTA,
    Cesium.Color.DARKOLIVEGREEN,
    Cesium.Color.DARKORANGE,
    Cesium.Color.DARKORCHID,
    Cesium.Color.DARKRED,
    Cesium.Color.DARKSALMON,
    Cesium.Color.DARKSEAGREEN,
    Cesium.Color.DARKSLATEBLUE,
    Cesium.Color.DARKSLATEGRAY,
    Cesium.Color.DARKTURQUOISE,
    Cesium.Color.DARKVIOLET,
    Cesium.Color.DEEPPINK,
    Cesium.Color.DEEPSKYBLUE,
    Cesium.Color.DIMGRAY,
    Cesium.Color.DODGERBLUE,
    Cesium.Color.FIREBRICK,
    Cesium.Color.FLORALWHITE,
    Cesium.Color.FORESTGREEN,
    Cesium.Color.FUCHSIA,
    Cesium.Color.GAINSBORO,
    Cesium.Color.GHOSTWHITE,
    Cesium.Color.GOLD,
    Cesium.Color.GOLDENROD,
    Cesium.Color.GRAY,
    Cesium.Color.GREEN,
    Cesium.Color.GREENYELLOW,
    Cesium.Color.HONEYDEW,
    Cesium.Color.HOTPINK,
    Cesium.Color.INDIANRED,
    Cesium.Color.INDIGO,
    Cesium.Color.IVORY,
    Cesium.Color.KHAKI,
    Cesium.Color.LAVENDER,
    Cesium.Color.LAVENDAR_BLUSH,
    Cesium.Color.LAWNGREEN,
    Cesium.Color.LEMONCHIFFON,
    Cesium.Color.LIGHTBLUE,
    Cesium.Color.LIGHTCORAL,
    Cesium.Color.LIGHTCYAN,
    Cesium.Color.LIGHTGOLDENRODYELLOW,
    Cesium.Color.LIGHTGRAY,
    Cesium.Color.LIGHTGREEN,
    Cesium.Color.LIGHTPINK,
    Cesium.Color.LIGHTSEAGREEN,
    Cesium.Color.LIGHTSKYBLUE,
    Cesium.Color.LIGHTSLATEGRAY,
    Cesium.Color.LIGHTSTEELBLUE,
    Cesium.Color.LIGHTYELLOW,
    Cesium.Color.LIME,
    Cesium.Color.LIMEGREEN,
    Cesium.Color.LINEN,
    Cesium.Color.MAGENTA,
    Cesium.Color.MAROON,
    Cesium.Color.MEDIUMAQUAMARINE,
    Cesium.Color.MEDIUMBLUE,
    Cesium.Color.MEDIUMORCHID,
    Cesium.Color.MEDIUMPURPLE,
    Cesium.Color.MEDIUMSEAGREEN,
    Cesium.Color.MEDIUMSLATEBLUE,
    Cesium.Color.MEDIUMSPRINGGREEN,
    Cesium.Color.MEDIUMTURQUOISE,
    Cesium.Color.MEDIUMVIOLETRED,
    Cesium.Color.MIDNIGHTBLUE,
    Cesium.Color.MINTCREAM,
    Cesium.Color.MISTYROSE,
    Cesium.Color.MOCCASIN,
    Cesium.Color.NAVAJOWHITE,
    Cesium.Color.NAVY,
    Cesium.Color.OLDLACE,
    Cesium.Color.OLIVE,
    Cesium.Color.OLIVEDRAB,
    Cesium.Color.ORANGE,
    Cesium.Color.ORANGERED,
    Cesium.Color.ORCHID,
    Cesium.Color.PALEGOLDENROD,
    Cesium.Color.PALEGREEN,
    Cesium.Color.PALETURQUOISE,
    Cesium.Color.PALEVIOLETRED,
    Cesium.Color.PAPAYAWHIP,
    Cesium.Color.PEACHPUFF,
    Cesium.Color.PERU,
    Cesium.Color.PINK,
    Cesium.Color.PLUM,
    Cesium.Color.POWDERBLUE,
    Cesium.Color.PURPLE,
    Cesium.Color.RED,
    Cesium.Color.ROSYBROWN,
    Cesium.Color.ROYALBLUE,
    Cesium.Color.SADDLEBROWN,
    Cesium.Color.SALMON,
    Cesium.Color.SANDYBROWN,
    Cesium.Color.SEAGREEN,
    Cesium.Color.SEASHELL,
    Cesium.Color.SIENNA,
    Cesium.Color.SILVER,
    Cesium.Color.SKYBLUE,
    Cesium.Color.SLATEBLUE,
    Cesium.Color.SLATEGRAY,
    Cesium.Color.SNOW,
    Cesium.Color.SPRINGGREEN,
    Cesium.Color.STEELBLUE,
    Cesium.Color.TAN,
    Cesium.Color.TEAL,
    Cesium.Color.THISTLE,
    Cesium.Color.TOMATO,
    Cesium.Color.TURQUOISE,
    Cesium.Color.VIOLET,
    Cesium.Color.WHEAT,
    Cesium.Color.WHITE,
    Cesium.Color.WHITESMOKE,
    Cesium.Color.YELLOW,
    Cesium.Color.YELLOWGREEN
];

var colorI = 0;

function nextColor() {
    c = colors[colorI];
    colorI += 1;
    if (colorI >= colors.length) {
        colorI = 0;
    }
    return c;
}
