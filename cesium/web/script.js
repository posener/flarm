// Consts from template.
const cameraLong = {{.Camera.Long }};
const cameraLat = {{.Camera.Lat }};
const cameraAlt = {{.Camera.Alt }};
const cameraHeading = {{.Camera.Heading }};
const cameraPitch = {{.Camera.Pitch }};

const altFix = {{.AltFix }};
const pathLength = {{.PathLength }};
const minGroundSpeed = {{.MinGroundSpeed }};
const units = "{{.Units }}";


// Your access token can be found at: https://cesium.com/ion/tokens.
Cesium.Ion.defaultAccessToken = '{{.Token}}';
// Initialize the Cesium Viewer in the HTML element with the `cesiumContainer` ID.
//
const viewer = new Cesium.Viewer('cesiumContainer', {
    terrainProvider: Cesium.createWorldTerrain(),
    imageryProvider: new Cesium.OpenStreetMapImageryProvider({url : 'https://a.tile.openstreetmap.org/'}),
    shadows: false,
    animation: false,
//    timeline: false,	
    geocoder: false,	
    homeButton: false,
    fullscreenButton: false,
    sceneModePicker: false,
    navigationHelpButton: false,
    requestRenderMode: true,
    maximumRenderTimeChange: 0.25
});

// Add Cesium OSM Buildings, a global 3D buildings layer.
viewer.scene.primitives.add(Cesium.createOsmBuildings());
//viewer.scene.debugShowFramesPerSecond = true;
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


function drawLLMGCTR(){
//Southern boundary
        viewer.entities.add({
            position: Cesium.Cartesian3.fromDegrees(35.22903, 32.59705, 70.0),
            ellipsoid: {
                radii: new Cesium.Cartesian3(3000.0, 3000.0, 10000.0),
                minimumClock: Cesium.Math.toRadians(0.0),
                maximumClock: Cesium.Math.toRadians(-180.0),
                minimumCone: Cesium.Math.toRadians(90.0),
                maximumCone: Cesium.Math.toRadians(88.3),
                material: Cesium.Color.BLACK.withAlpha(0.2),
            },
        });

//Northern boundary
        viewer.entities.add({
            position: Cesium.Cartesian3.fromDegrees(35.22903, 32.59705, 70.0),
            ellipsoid: {
                radii: new Cesium.Cartesian3(1800.0, 1800.0, 10000.0),
                minimumClock: Cesium.Math.toRadians(0.0),
                maximumClock: Cesium.Math.toRadians(180.0),
                minimumCone: Cesium.Math.toRadians(90.0),
                maximumCone: Cesium.Math.toRadians(88.48),
                material: Cesium.Color.BLACK.withAlpha(0.2),
            },
        });

//Middle boundary
        viewer.entities.add({
            position: Cesium.Cartesian3.fromDegrees(35.22903, 32.59705, 182.88),
            plane: {
                plane: new Cesium.Plane(Cesium.Cartesian3.UNIT_Y, 0.0),
                dimensions: new Cesium.Cartesian2(6000.0, 365.76),
                material: Cesium.Color.BLACK.withAlpha(0.2),
            },
        });
}

function drawCircuit(){
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
}


function drawRamatDavidCTR(){
var d = new Date();
var n = d.getDay()

if(n == 5 || n == 6){
		//RAMAT DAVID CTR SHRINKED
		viewer.entities.add({
			name: "RAMAT DAVID CTR SHRINKED",
			wall: {
			positions: Cesium.Cartesian3.fromDegreesArrayHeights([
			//Turkish Bridge
		        35.10805556,
			32.72194444,
			3657.6,
		        //Zarzir
                	35.20861111,
		        32.73083333,
		        3657.6,
		        //Tzomet Adashim
		        35.29222222,
		        32.65861111,
		        3657.6,
		        //Balfouria
		        35.29000000,
		        32.62894444,
		        3657.6,
		        //Ha-Yogev
		        35.20222222,
		        32.61472222,
		        3657.6,
		        //Mishmar Ha-Emeq
		        35.14166667,
		        32.61527778,
		        3657.6,
		        //Hazorea
		        35.12388889,
		        32.64444444,
		        3657.6,
		        //Turkish Bridge
		        35.10805556,
		        32.72194444,
		        3657.6,
			]),
			material: Cesium.Color.DARKRED.withAlpha(0.1),
			outline: true,
			outlineColor: Cesium.Color.BLACK,
			},
		});
}else{
		//RAMAT DAVID CTR FULL
		viewer.entities.add({
			name: "RAMAT DAVID CTR FULL",
			wall: {
			positions: Cesium.Cartesian3.fromDegreesArrayHeights([
			//Kfar Hasidim
			35.092222,
			32.749444,
			3657.6,
			//Tzomet Ha-Movil
			35.233611,
			32.757778,
			3657.6,
			//Natzeret
			35.300278,
			32.705000, 
			3657.6,
			//Afula
			35.289167,
			32.607222,
			3657.6,
			//Tzomet Megido
			35.193333,
			32.572778,
			3657.6,
			//Ein HaShofet
			35.092778,
			32.598611,
			3657.6,
			//Kfar Hasidim
			35.092222,
			32.749444,
			3657.6,
			]),
			material: Cesium.Color.RED.withAlpha(0.1),
			outline: true,
			outlineColor: Cesium.Color.BLACK,
			},
		});
}
}

function drawRoutes(){
	//J14
	var orangeOutlined = viewer.entities.add({
	  name:"J14 route",
	  polyline: {
		positions: Cesium.Cartesian3.fromDegreesArrayHeights([
			35.05666,
			32.41138,
			1219.2,
			35.25527,
			32.58555,
			1219.2,
			35.54333,
			32.83638,
			1219.2,
	    	]),
	width: 5,
	material: new Cesium.PolylineOutlineMaterialProperty({
		color: Cesium.Color.ORANGE,
		outlineWidth: 2,
		outlineColor: Cesium.Color.BLACK,
    	}),
  	},
	});
}

function drawModel(airplaneType) {
    switch (airplaneType) {
        case "glider":
            return new Cesium.ModelGraphics({
                uri: "/models/glider/scene.gltf",
                allowPicking: 1,
                minimumPixelSize: 6,
            });
	break;
        case "towplane":
            return new Cesium.ModelGraphics({
                uri: "/models/towplane/towplane.gltf",
                allowPicking: 1,
                minimumPixelSize: 18,
            });
	break;
        default:
            return new Cesium.ModelGraphics({
                uri: "/models/ufo/scene.gltf",
                allowPicking: 1,
		minimumPixelSize: 6,
		scale: 0.1,
            });
    }
}
function setMarkerText(msg){
	var marker = `${msg.Name}\n`;
	switch(units){
		case "metric":
			marker = marker +
      			`Alt: ${msg.Alt}m\n` +
        		`GS: ${msg.GroundSpeed}m/s\n` +
        		`Climb: ${msg.Climb}m/s\n` +
        		`Heading: ${msg.Dir}°`;
		break;
		case "imperial":
			 marker = marker +
                        `Alt: ${msg.Alt}ft\n` +
                        `GS: ${msg.GroundSpeed}kts\n` +
                        `Climb: ${msg.Climb}kt/s\n` +
                        `Heading: ${msg.Dir}°`;
		break;
                case "mixed":
                         marker = marker +
                        `Alt: ${Math.round(msg.Alt*3.28084)}ft\n` +
                        `GS: ${Math.round(msg.GroundSpeed*1.94384)}kts\n` +
                        `Climb: ${msg.Climb}m/s\n` +
                        `Heading: ${msg.Dir}°`;
	}
	return marker;
}

function main() {
    drawLLMGCTR();
    drawCircuit();
    drawRamatDavidCTR();

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
        const id = msg.Name;
//Correction for true heading
        var heading = Cesium.Math.toRadians(msg.Dir-90.0);
        var pitch = Cesium.Math.toRadians(0.0);
        var roll = Cesium.Math.toRadians(0.0);
        var orientation = Cesium.Transforms.headingPitchRollQuaternion(position, new Cesium.HeadingPitchRoll(heading, pitch, roll));

        if (msg.GroundSpeed < minGroundSpeed) {
            return;
        }
	
        if (!viewer.entities.getById(id)) {
            console.log(`Creating ${id}.`);
            var posProp = new Cesium.SampledPositionProperty();
            viewer.entities.add({
                id: id,
                position: posProp,
                model: drawModel(msg.Type),
                description: `${id}`,
                //orientation: new Cesium.VelocityOrientationProperty(posProp),
                path: new Cesium.PathGraphics({
                    width: 2,
		    leadTime: 0,
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
        entity.position.setInterpolationOptions({
          interpolationDegree: 1,
          interpolationAlgorithm: Cesium.LagrangePolynomialApproximation,
        });
        entity.orientation = orientation;

        entity.label = {
            text: setMarkerText(msg),
            font: '16pt monospace',
	    fillColor: Cesium.Color.BLACK,
	    horizontalOrigin: Cesium.HorizontalOrigin.LEFT,
            pixelOffset: new Cesium.Cartesian2(0, -50),
            scaleByDistance: new Cesium.NearFarScalar(0.0, 1.0, 1.0e4, 0.5)
        };
//	viewer.scene.requestRender();
    };
    ws.onclose = function () {
        console.log(`ws disconnected`);
    };
}

main();

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

function nextColor() {
    var c = colors[Math.floor(Math.random() * 138)+1];
    return c;
}
