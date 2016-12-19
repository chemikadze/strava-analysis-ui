function drawGraph(data) {
    return scatterPlotCustom(data, {
        calcY: function(d) {
            return d.total_elevation_gain;
        },
        titleY: "Elevation gain, m",
    });
}