function drawGraph(data) {
    return scatterPlotCustom(data, {
        calcY: function(d) {
            return d.elapsed_time / 60 / 60;
        },
        titleY: "Elapsed time, hr",
    });
}