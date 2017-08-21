var weekLength = 1*60*60*24*7*1000;

function weekStart(timestamp) {
    return weekIndex(timestamp) * weekLength;
}

function weekIndex(timestamp) {
  return Math.floor(timestamp / weekLength);
}

function trimByPredicate(array, predicate) {
  if (array.length == 0) {
    return array;
  }
  var first = 0;
  var last = array.length - 1;
  for (var i = 0; i < array.length; i++) {
    if (predicate(array[i])) {
      break;
    }
  }
  first = i;
  for (var i = last; i >= 0; i--) {
    if (predicate(array[i])) {
      break;
    }
  }
  return array.slice(first, last+1);
}

function drawGraph(data) {
    var parseTime = d3.isoParse;
    data.reverse();
    var minWeek = weekIndex(parseTime(data[0].start_date));
    var maxWeek = weekIndex(parseTime(data[data.length - 1].start_date));
    var weeklySufferScore = d3.nest()
      .key(function(d) { return weekStart(parseTime(d.start_date)); })
      .rollup(function(v) {
        return v.map(function (x) { return x.suffer_score; })
            .reduce(function (x, y) { return x + y; });
      })
      .entries(data);
    var entriesByKey = {};
    for (var i = 0; i < weeklySufferScore.length; i++) {
      entriesByKey[weeklySufferScore[i].key] = weeklySufferScore[i];
    }
    var weeklySufferScoreWithoutHoles = [];
    for (var i = minWeek; i <= maxWeek; i++) {
      var key = (i * weekLength).toString();
      var value = entriesByKey[key];
      if (value && value.value) {
        weeklySufferScoreWithoutHoles.push(value);
      } else {
        weeklySufferScoreWithoutHoles.push({key: key, value: 0});
      }
    }
    weeklySufferScoreWithoutHoles = trimByPredicate(weeklySufferScoreWithoutHoles, function (d) { return d.value; } );
    return barPlotCustom(weeklySufferScoreWithoutHoles, {
        calcX: function(d) { return d.key; },
        calcY: function(d) { return d.value; },
        titleY: "Weekly suffer score",
        calcTooltip: function(d) { return parseTime(parseInt(d.key)).toDateString(); },
    });
}