function scatterPlotCustom(data, meta) {

    var predicate = meta.predicate || function(x) { return true };
    var calcX = meta.calcX || function(d) { return parseTime(d.start_date); }
    var calcY = meta.calcY;
    var titleY = meta.titleY || "";

    data = data.filter(predicate);

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
        .style("fill", function(d) { return "#FF0000"; })
        .on('click', function(d) { window.open("https://www.strava.com/activities/" + d.id); }, true);
}