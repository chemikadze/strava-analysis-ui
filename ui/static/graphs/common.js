function getListOfClusters(data) {
    var clusters = [];
    for (var i = 0; i < data.length; i++) {
        var pos = clusters.indexOf(data[i]);
        if (pos < 0) {
            clusters.push([data[i], 1]);
        } else {
            clusters[pos][1] += 1;
        }
    }
    clusters.sort(function (a, b) {
        return b[1] - a[1]; // descending order
    });
    return clusters.map(function (a) {
        return a[0];
    });
}

function getClusterId(clusterFeature, clusters) {
    return clusters.indexOf(clusterFeature);
}

function getClusterColor(clusterId, colors) {
    return colors[clusterId % colors.length];
}

function scatterPlotCustom(data, meta) {

    var predicate = meta.predicate || function(x) { return true };
    var calcX = meta.calcX || function(d) { return parseTime(d.start_date); }
    var calcY = meta.calcY;
    var titleY = meta.titleY || "";
    var clusterBy = meta.clusterBy || function(d) { return null; }
    var clusterColors = ["#FF0000", "#00D200", "#0000FF", "#FF00FF", "#00FFFF", "#FFFF00"];

    data = data.filter(predicate);

    var clusters = getListOfClusters(data.map(clusterBy));

    var parseTime = d3.isoParse;  

    var svg = d3.select("svg"),
        margin = {top: 20, right: 20, bottom: 30, left: 50},
        width = +svg.attr("width") - margin.left - margin.right,
        height = +svg.attr("height") - margin.top - margin.bottom,
        g = svg.append("g").attr("transform", "translate(" + margin.left + "," + margin.top + ")");    

    var x = d3.scaleTime()
        .rangeRound([0, width]);

    var y = d3.scaleLinear()
        .rangeRound([height, 0]);

    x.domain(d3.extent(data, function(d) { return calcX(d); }));
    y.domain(d3.extent(data, function(d) { return calcY(d); }));

    var lineFunction = d3.line()
                         .x(function(d) { return x(calcX(d)); })
                         .y(function(d) { return y(calcY(d)); });

    g.append("g")
        .attr("class", "axis axis--x")
        .attr("transform", "translate(0," + height + ")")
        .call(d3.axisBottom(x));

    g.append("g")
        .attr("class", "axis axis--y")
        .call(d3.axisLeft(y))
      .append("text")
        .attr("fill", "#000")
        .attr("transform", "rotate(-90)")
        .attr("y", 6)
        .attr("dy", "0.71em")
        .style("text-anchor", "end")
        .text(titleY);

    g.selectAll(".dot")
        .data(data)
      .enter().append("circle")
        .attr("class", "dot")
        .attr("cx", function(d) { return x(calcX(d)); })
        .attr("cy", function(d) { return y(calcY(d)); })
        .style("fill", function(d) { return getClusterColor(getClusterId(clusterBy(d), clusters), clusterColors); })
        .on('click', function(d) { window.open("https://www.strava.com/activities/" + d.id); }, true);
}