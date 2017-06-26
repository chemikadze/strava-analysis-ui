function drawGraph(data) {
    return scatterPlotCustom(data, {
        calcY: function(d) {
            return d.distance / 1000 / (d.moving_time / 60 / 60);
        },
        titleY: "Elapsed time, hr",
        clusterBy: function(d) {
            return d.gear_id;
        }
    });
}    