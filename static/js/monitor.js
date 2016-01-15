console.log("Staring Monitor.js");
var url=window.location.href 
console.log(url)
var myPattern="/accounts/([a-zA-Z0-9]+)/venues/([a-zA-Z0-9]+)/stocks/([a-zA-Z0-9]+)$"

var matches= url.match(myPattern);
console.log(matches)

account= matches[1]
console.log("Account:"+account);
venue= matches[2]
console.log("Venue:"+venue);
stock= matches[3]
console.log("Stock:"+stock);

var quotesArray = [];
var format = d3.time.format("%Y-%m-%dT%H:%M:%S");

var margin = {top: 20, right: 20, bottom: 30, left: 50},
    width = 960 - margin.left - margin.right,
    height = 500 - margin.top - margin.bottom;

var formatDate = d3.time.format("%d-%b-%y");

var xTime = d3.time.scale()
    .range([0, width]);

var xIndex = d3.scale.linear()
    .range([0, width]);

var y = d3.scale.linear()
    .range([height,0]);


var xTimeAxis = d3.svg.axis()
    .scale(x)
    .orient("bottom")
    .tickFormat(d3.time.format("%Hh %Mm %Ss"))
    .ticks(d3.time.seconds,5);

var xIndexAxis = d3.svg.axis()
    .scale(xIndex)
    .orient("bottom")
    .ticks(5);

var yAxis = d3.svg.axis()
    .scale(y)
    .orient("left")
    .ticks(5);

var x=xIndex;
var xAxis=xIndexAxis;


var chartSvg = d3.select("body").append("svg")
    .attr("width", width + margin.left + margin.right)
    .attr("height", height + margin.top + margin.bottom)
    .append("g")
    .attr("transform", "translate(" + margin.left + "," + margin.top + ")");

    chartSvg.append("g")
    .attr("class", "x axis")
    .attr("transform", "translate(0," + height + ")");


chartSvg.append("g")
    .attr("class", "y axis")
    .call(yAxis)
    .append("text")
    .attr("transform", "rotate(-90)")
    .attr("y", 6)
    .attr("dy", ".71em")
    .style("text-anchor", "end")
    .text("Price ($)");

var textSvg = d3.select("body").append("svg")
    .attr("width", width + margin.left + margin.right)
    .attr("height", height + margin.top + margin.bottom);


var line = d3.svg.line()
    .x(function(d) { return x(d.index); })
    .y(function(d) { return y(d.price/100); })
    .interpolate("linear");

var area= d3.svg.area()
    .x(function(d) { return x(d.quoteTime); })
    .y0(height)
    .y1(function(d) { return y(d.price/100); });

chartSvg.append("path")
    .attr("class", "line")
    .attr("stroke", "black")
    .attr("stroke-width", 1)
    .attr("fill", "none");

var quoteCounter = 0;
var quotesWebSocker = new WebSocket("wss://api.stockfighter.io/ob/api/ws/"+account+"/venues/"+venue+"/tickertape/stocks/"+stock);


quotesWebSocker.onmessage = function (event) {
    var ticker = JSON.parse(event.data);
    quoteCounter++;
    var quote = { "quoteTime":format.parse(ticker.quote.quoteTime.slice(0,19)), "price":ticker.quote.last, "index":quoteCounter};

    if (quotesArray.length > 0 && ((new Date(quote.quoteTime)).getTime() > (new Date(quotesArray[quotesArray.length])).getTime())) {
        quotesArray.push(quote);
    } else {
        quotesArray.push(quote);
    }
}

var tick = 0;

var graphQuoteArray = [];
for (var i = 0; i < width; i++) {
        graphQuoteArray.push({ "quoteTime":0, "price":0, "index":-width+i+1});
}


/*
setInterval(function() {
    console.log("Loop cycle: "+(tick++));
    /*
    if (quotesArray.length > 0) {
        console.log("Data in");

        var diff =   quotesArray.length  - width;
        if (diff > 0) {
            graphQuoteArray = quotesArray.slice(quotesArray.length-width,quotesArray.length);
            chartSvg.select(".line")
            .attr("transform", "translate(" + (-diff) + ",0)");
        } else {
            graphQuoteArray = quotesArray.slice(0,quotesArray.length);
        }

        chartSvg.select("path")
        .transition();

        chartSvg.select(".line")
        .attr("d",line(graphQuoteArray));

        var xExtent = d3.extent(graphQuoteArray, function(quote){return  quote.index;}) 
        x.range([0,width])
        x.domain(xExtent);
        chartSvg.selectAll("g.x.axis").call(xAxis);


        var yExtent = d3.extent(graphQuoteArray, function(quote) { return quote.price; });

        y.domain([yExtent[0]/100-20,yExtent[1]/100+20]);
        chartSvg.select("g.y.axis").call(yAxis);
    }

    if (quotesArray.length > 0) {

        var lastGraphIndex =  graphQuoteArray[graphQuoteArray.length-1].index;
        var lastQuoteIndex =  quotesArray[quotesArray.length-1].index;

        if (lastGraphIndex < lastQuoteIndex) { 
            graphQuoteArray = graphQuoteArray.concat(quotesArray.slice(lastGraphIndex,quotesArray.length));
        } 
        var diff =  graphQuoteArray.length - width;

        //x axis update

        //var tempSvg = d3.select("body").transition();



        //console.log (graphQuoteArray.length);


        //y axis update

        //tempSvg.select(".y.axis")
        //.call(yAxis);

        //tempSvg.select(".x.axis")
        //.call(xAxis)
        //.duration(1000)
        //.attr("transform", "translate(" + x(-diff) + ",0)");
        
        var xExtent = d3.extent(graphQuoteArray, function(quote){return quote.index;}) 
        x.domain(xExtent);

        var yExtent = d3.extent(graphQuoteArray, function(quote) { return quote.price; });
        y.domain([yExtent[0]/100-20,yExtent[1]/100+20]);

        chartSvg.select(".line")
        .attr("d",line(graphQuoteArray))
        .attr("transform",null);

        //path update
        //chartSvg.select(".area").attr("d",area(graphQuoteArray));
        

        console.log (x(diff));
        console.log (x(-diff));


        if (diff > 0) {
            for (var i=0; i< diff ; i++) {
                console.log ("diff > 0");
                chartSvg.select(".line")
                .transition()
                .duration(200)
                .attr("d",line(graphQuoteArray))
                .attr("transform", "translate(-" + i + ",0)");
            }

            graphQuoteArray = graphQuoteArray.slice(diff, graphQuoteArray.length);
        }



           var xExtent = d3.extent(graphQuoteArray, function(quote){return  quote.quoteTime;}) 
           if (graphQuoteArray.length < width) {
           x.range([0,graphQuoteArray.length])
           x.domain(xExtent);
           } else {
           x.range([0,width])
           x.domain([xExtent[1]-x(graphQuoteArray[graphQuoteArray.length-width].quoteTime),xExtent[1]]);
           }

           chartSvg.selectAll("g.x.axis").call(xAxis);

        //Draw text labels and value 
           var textArray= [{"label":"Price: ","value":quote.price},{"label":"Quote Time: ","value":quote.quoteTime} ]

           var text = textSvg.selectAll("text")
           .data(textArray);

           text.exit().remove();


           text.enter().append("text")
           .attr("x", 20)
           .attr("y", function(d,i) { return i*30 + 20; })
           .attr("font-family", "sans-serif")
           .attr("font-size", "20px");

           text.text(function(d) {return d.label + d.value;});
        } 
        }
}, 1000);

*/
var duration = 0;
var durationToggle = 1;
var counter = 0;
var avg = 0;


function updateGraph() {
    //console.log("Loop cycle: "+(tick++));


    if (quotesArray.length > 0) {

        var lastGraphIndex =  graphQuoteArray[graphQuoteArray.length-1].index;
        var lastQuoteIndex =  quotesArray[quotesArray.length-1].index;

        if (lastGraphIndex < lastQuoteIndex) { 
            graphQuoteArray[graphQuoteArray.length] = quotesArray[lastGraphIndex];


            var xExtent = d3.extent(graphQuoteArray, function(quote){return quote.index;}) 
            x.domain(xExtent);

            var yExtent = d3.extent(graphQuoteArray, function(quote) { return quote.price; });
            y.domain([yExtent[0]/100-(yExtent[0]/100)*0.10,yExtent[1]/100+(yExtent[1]/100)*0.10]);

            chartSvg.select(".line")
            .attr("d",line(graphQuoteArray))
            .transition()
            .duration(duration)
            .attr("transform",null);

            chartSvg.select("g.x.axis").call(xAxis);
            chartSvg.select("g.y.axis").call(yAxis);


            graphQuoteArray.shift();
        } 
    } 
    if (lastQuoteIndex - lastGraphIndex< 200) {
        interval = 40;
    } else {
        interval = 10;
    }
    setTimeout(updateGraph, interval);
}

setTimeout(updateGraph, 100);
