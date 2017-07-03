function drawGraph(data) {
    return scatterPlotCustom(data, {
        titleY: "Average power, W",
        calcY: function(d) {
            return d.average_watts;
        },
        clusterBy: function(d) {
            return d.gear_id;
        }
    });
}