function drawGraph(data) {
    return scatterPlotCustom(data, {
        calcY: function(d) {
            return d.average_speed / d.average_heartrate;
        },
        titleY: "Avg speed per bpm",
        predicate: function(d) {
            return !!d.average_heartrate;
        },
        clusterBy: function(d) {
            return d.gear_id;
        }
    });
}