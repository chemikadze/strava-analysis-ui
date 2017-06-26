function drawGraph(data) {
    return scatterPlotCustom(data, {
        calcY: function(d) {
            return d.average_watts / d.average_heartrate;
        },
        titleY: "Avg power per bpm (including estimates)",
        predicate: function(d) {
            return !!d.average_watts && !!d.average_heartrate;
        },
        clusterBy: function(d) {
            return d.gear_id;
        }
    });
}