function drawGraph(data) {
    return scatterPlotCustom(data, {
        calcY: function(d) {
            return d.distance;
        },
        titleY: "Distance, km",
    });
}