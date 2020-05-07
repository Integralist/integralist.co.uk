---
title: "Datadog Metric Searcher"
date: 2020-05-01T09:29:41+01:00
categories:
  - "code"
  - "guide"
tags:
  - "bash"
  - "python"
  - "datadog"
  - "metrics"
  - "apis"
draft: false
---

- [We're over our data limit!](#we-re-over-our-data-limit)
- [Calculating Custom Metrics](#calculating-custom-metrics)
- [Tracking Metrics](#tracking-metrics)
- [Automation FTW](#automation-ftw)
- [`DISTRIBUTION` metric type?](#distribution-metric-type)
  - [Metric Types](#metric-types)
  - [Multiple Metrics](#multiple-metrics)
  - [Percentile Aggregations](#percentile-aggregations)
  - [Tag Filtering](#tag-filtering)
- [Choosing `HISTOGRAM` or `DISTRIBUTION` ?](#choosing-histogram-or-distribution)
  - [The `DISTRIBUTION` percentile 'custom metric' difference](#the-distribution-percentile-custom-metric-difference)

## We're over our data limit!

I've spent quite a bit of time on metrics the past month. Our contract with [Datadog](https://www.datadoghq.com/) was coming up for renewal and the feedback we had received was that our organization were vastly over our allotted data limits and we needed to address this in time for the contract renewal discussions.

> **Note**: Datadog has been awesome when it comes to these types of discussions and have been very flexible with us.

NO, this post is not sponsored by Datadog. 

I wish it was yer know `:make-it-rain:`

![make it rain](../../images/make-it-rain.webp).

The solution to our problem was a multi-pronged approach, and what I'm going to focus on in this post is one small aspect of that. 

Before we get into that, let's first take a moment to understand how costs are calculated. Datadog has various [pricing structures](https://docs.datadoghq.com/account_management/billing/pricing/) and one of the items it considers are the number of 'custom metrics' that your service(s) are generating.

**So what constitutes a 'custom metric'?** Let's find out...

## Calculating Custom Metrics

The amount of metrics you 'report' does not equate to the same thing as the number of 'custom metrics' you are billed for.

> A custom metric is uniquely identified by a combination of a metric name and tag values (including the host tag). — [Datadog](https://docs.datadoghq.com/account_management/billing/custom_metrics/)

Consider a web server application that reports the time it takes for an incoming request to be served back to the client. Imagine we report this metric (`request.latency`) with no [tags](https://docs.datadoghq.com/tagging/). Reporting this metric would result in us being charged for a single 'custom metric'. 

Now imagine we report `request.latency` with three tags (`host`, `endpoint`, `status`), then with the following combination of unique tag values, we'd end up with _four_ custom metrics:

![custom-metric](../../images/custom-metric.png)

The situation is made worse when using a `HISTOGRAM` metric type, as it ultimately produces five separate metrics. Meaning the `request.latency` example, if reported as a `HISTOGRAM`, would result in twenty custom metrics (`5 metrics * 4 tag combinations`). 

If you were to add the URL request path as a tag to the `request.latency` metric, then you can imagine that having quite a 'high-cardinality' (i.e. a large number of variants) depending on the number of endpoints that service was handling.

This demonstrates how we must be careful with the cardinality of any tags we apply to our metrics.

## Tracking Metrics

Datadog provides many tools to help us track down expensive metrics, but one particular problem we had was identifying whether the metrics we were reporting were actually being used. For example, were our metrics being referenced in a monitor or a dashboard? If not, they could be deleted.

This was considered a 'low-hanging' fruit task, but I found the journey to my solution to be quite interesting and so I wanted to share it in case it was of interest to y'all too.

Datadog's UI provides us with tools to _manually_ check what metrics appear within a specified timeframe, but (at the time of writing) we have nearly 600+ microservices all producing a large volume of metrics, and we also have near to 1000 dashboards and 1000 monitors. 

I needed a way to automate things, and so this is where Datadog's developer APIs came to the rescue.

## Automation FTW

So we needed a way to identify what metrics we were actually using. I decided to utilise Datadog's developer APIs to help me.

> **Caveat**: this isn't a perfect solution, and I offer no guarantees to its accuracy, but what I can say is that this not only helped reduce my workload _dramatically_ but also helped us to identify potential cases where we could switch from either a HISTOGRAM/TIMER metric type over to a DISTRIBUTION metric type which in itself was another avenue to help us reduce costs (I'll add information about this later in the post).

Below is the output of the script I wrote (yes, I've generalized the output to protect the guilty). In there we have specific sections, such as...

- METRICS TO FIND
- DASHBOARDS
- UNUSED METRICS IN DASHBOARDS
- MONITORS
- UNUSED METRICS IN MONITORS
- UNUSED METRICS IN MONITORS AND DASHBOARDS

Pretty much all of this output, with the exception of the last section, can be turned off with a flag. I typically like to print all the information (or to `tee` it off into backup files) for contextual and informational purposes. But other people using the script likely don't care and just want to know what metrics can be deleted.

```
###############
METRICS TO FIND
###############

[
  {
    "name": "namespace.foo.bar",
    "count": 0
  },
  {
    "name": "namespace.baz.qux",
    "count": 0
  },
  {
    "name": "namespace.beep.boop",
    "count": 0
  },
] 


no search pattern provided, meaning we'll search ALL <N> dashboards and <N> monitors!

##########
DASHBOARDS
##########

My Foo Service Dashboard Title

https://example.datadoghq.com/dashboard/xyz-abc-123/foo-service

{
  "foo bar": [
    {
      "metric": "avg:namespace.foo.bar{$scope}.as_count()",
      "type": "timeseries"
    }
  ],
} 

---------

############################
UNUSED METRICS IN DASHBOARDS
############################

namespace.baz.qux
namespace.beep.boop

########
MONITORS
########

My Foo Service Monitor 

https://example.datadoghq.com/monitors/1234567 

min(last_5m):avg:namespace.baz.qux.95percentile{environment:prod} by {handler} > <N> 

---------

##########################
UNUSED METRICS IN MONITORS
##########################

namespace.foo.bar
namespace.beep.boop

#########################################
UNUSED METRICS IN MONITORS AND DASHBOARDS
#########################################

namespace.beep.boop

1 out of 3 metrics are unused.
```

So now we know what the output of the script is, let's dig into the code itself. I won't explain every single line, but I will pick out interesting segments. It's worth noting that my original implementation was not asynchronous and so the runtime execution was ~5-6 minutes depending on the size of the dashboards we have (e.g. how many graphs appear within a single dashboard).

I later refactored the code (which is the version you'll be looking at) and this helped to drop the runtime down to ~20 seconds. Below is the script, which is written using Python 3.8.

```
import argparse
import concurrent.futures
import json
import operator
import re

from datadog import api, initialize

options = {
    "api_key": "foo",
    "app_key": "foo"
}

initialize(**options)


def pprint(o):
    """pretty print data structures."""
    print(json.dumps(o, indent=2, default=str), "\n")


def format_title(t):
    """print title in a format that makes it stand out visually.
    example: "my title" > "\n########\nMY TITLE\n########\n"
    """
    hashes = "#" * len(t)
    print(f"\n{hashes}\n{t.upper()}\n{hashes}\n")


parser = argparse.ArgumentParser(
    description="Datadog Metric Searcher (searches dashboards and monitors)")

parser.add_argument(
    "-p", "--pattern", default=".",
    help="regex pattern for filtering dashboard/monitor by name")

parser.add_argument(
    "-m",
    "--metrics",
    help="comma-separated list of metrics",
    required=True)

parser.add_argument(
    "-u",
    "--unused",
    help="only display unused metrics",
    action="store_true")

args = parser.parse_args()

metrics = [{"name": metric, "count": 0}
           for metric in args.metrics.split(",") if metric]


def find_graphs(
        widgets,
        dashboard_title,
        dashboard_url,
        metrics,
        matches=None):
    """recursively search dashboard graphs for those referencing our metrics.

    Note: widgets can be nested multiple times, so this is a recursive function.

    because this function is run in isolation within its own process we pass in
    the dashboard title/url so we can report back within the main/parent process
    which dashboard the graphs are associated with (as the results are received
    based on which process is quickest to complete). we also pass in a list of
    metrics to be looked up in the dashboards/graphs, as we can't manipulate the
    metric list (defined in the parent process) from within a child process.
    """

    if not matches:
        matches = {}

    for widget in widgets:
        definition = widget.get("definition")

        if definition["type"] != "note":
            requests = definition.get("requests")

            """
            example data:
            {
              "style": {
                "palette": "green_to_orange",
                "palette_flip": false
              },
              "group": [],
              "title": "Hosts",
              "node_type": "container",
              "no_metric_hosts": true,
              "scope": [
                "$cluster",
                "rig.service:user_auth_proxy"
              ],
              "requests": {
                "fill": {
                  "q": "avg:process.stat.container.io.wbps{$cluster,rig.service:user_auth_proxy} by {host}"
                }
              },
              "no_group_hosts": true,
              "type": "hostmap"
            }
            """

            if requests:
                for request in requests:
                    metric_query = None
                    log_query = None
                    process_query = None

                    """
                    the following if statement catches 'hostmap' graphs
                    whose requests key is a dict, not list[dict]
                          "requests": {
                            "fill": {
                              "q": "..."
                            }
                          }
                    """
                    if isinstance(requests, dict) and request == "fill":
                        metric_query = requests.get(request, {}).get("q")
                    else:
                        try:
                            metric_query = request.get("q")
                        except AttributeError as err:
                            continue

                        log_query = request.get(
                            "log_query", {}).get(
                            "search", {}).get("query")
                        process_query = request.get(
                            "process_query", {}).get("metric")

                    if metric_query:
                        query = metric_query
                    elif log_query:
                        query = log_query
                    elif process_query:
                        query = process_query
                    else:
                        query = None

                    if not query:
                        continue

                    for metric in metrics:
                        if metric["name"] in query:
                            metric["count"] += 1

                            graph_title = definition.get("title", "N/A")
                            match = matches.get(graph_title)

                            if not match:
                                matches[graph_title] = []
                                match = matches[graph_title]

                            match.append({
                                "metric": query,
                                "type": definition["type"],
                            })
            else:
                nested_widgets = definition.get("widgets", [])

                # recurse and ignore the dashboard title/url and metrics
                # as from this stage of the function we don't care about them
                _, _, _, d = find_graphs(
                    nested_widgets,
                    dashboard_title,
                    dashboard_url,
                    metrics,
                    matches
                )

                matches.update(d)

    return dashboard_title, dashboard_url, metrics, matches


def all_dashboards():
    """acquire all dashboards."""

    return api.Dashboard.get_all()


def all_monitors():
    """acquire all monitors."""

    return api.Monitor.get_all()


def dashboard_get(dashboard: dict):
    """acquire dashboard by the given ID."""

    return api.Dashboard.get(dashboard["id"])


def filter_dashboards(dashboards):
    """filter dashboards by pattern provided by -p/--pattern flag."""

    filtered_dashboards = []

    for dashboard in dashboards["dashboards"]:
        if re.search(args.pattern, dashboard["title"], flags=re.IGNORECASE):
            filtered_dashboards.append(
                {
                    "title": dashboard["title"],
                    "id": dashboard["id"],
                    "url": dashboard["url"],
                }
            )

    return sorted(filtered_dashboards, key=operator.itemgetter("title"))


def filter_monitors(monitors):
    """filter monitors by pattern provided by -p/--pattern flag."""

    filtered_monitors = []

    for monitor in monitors:
        if re.search(args.pattern, monitor["name"], flags=re.IGNORECASE):
            filtered_monitors.append(
                {"name": monitor["name"],
                 "url": f"https://<YOUR_ORG>.datadoghq.com/monitors/{monitor['id']}",
                 "query": monitor["query"], })

    return sorted(filtered_monitors, key=operator.itemgetter("name"))


def process():
    """asynchronously acquire dashboards and update metric count.

    Note: the Datadog API is not asynchronous, so we must run API operations
    within a threadpool, while also running the metric 'searching' algorithm
    (a cpu heavy operation) within a processpool to help speed up the overall
    program execution time.
    """

    if not args.unused:
        format_title("metrics to find")
        pprint(metrics)

    dashboards = None
    monitors = None

    with concurrent.futures.ThreadPoolExecutor() as executor:
        wait_for = [
            executor.submit(all_dashboards),
            executor.submit(all_monitors)
        ]

        for f in concurrent.futures.as_completed(wait_for):
            """identify which api finished first and assign to correct variable.
            dashboard api returns a dictionary, while monitors returns a list.
            """

            results = f.result()

            if isinstance(results, dict):
                dashboards = results
            else:
                monitors = results

    if args.pattern == ".":
        ld = len(dashboards['dashboards'])
        lm = len(monitors)
        d = f"{ld} dashboards"
        m = f"{lm} monitors"
        msg = f"\nno search pattern provided, meaning we'll search ALL {d} and {m}!\n"
        print(msg)

    filtered_dashboards = filter_dashboards(dashboards)
    filtered_monitors = filter_monitors(monitors)

    dashboards_metadata = []
    track_dashboard_metrics = {}

    with concurrent.futures.ThreadPoolExecutor() as executor:
        wait_for = [
            executor.submit(dashboard_get, dashboard)
            for dashboard in filtered_dashboards
        ]

        for f in concurrent.futures.as_completed(wait_for):
            dashboards_metadata.append(f.result())

    if not args.unused:
        format_title("dashboards")

    with concurrent.futures.ProcessPoolExecutor() as executor:
        metrics_copy = metrics.copy()  # avoid accidental mutation
        wait_for = [
            executor.submit(
                find_graphs,
                dashboard["widgets"],
                dashboard["title"],
                dashboard["url"],
                metrics_copy) for dashboard in dashboards_metadata]

        for f in concurrent.futures.as_completed(wait_for):
            title, url, metrics_mod, matches = f.result()
            if matches:
                if not args.unused:
                    print(title, "\n")
                    print(f"https://<YOUR_ORG>.datadoghq.com{url}\n")
                    pprint(matches)
                    print("---------\n")

            for metric in metrics_mod:
                if not track_dashboard_metrics.get(metric["name"]):
                    track_dashboard_metrics.update({metric["name"]:
                                                    metric["count"]})
                else:
                    track_dashboard_metrics[metric["name"]] += metric["count"]

    unused_dashboard_metrics = set()
    unused_monitor_metrics = set()
    used_monitor_metrics = set()

    if not args.unused:
        format_title("unused metrics in dashboards")

    for metric, count in track_dashboard_metrics.items():
        if count == 0:
            unused_dashboard_metrics.add(metric)
            print(metric)

    if not args.unused:
        format_title("monitors")

    for monitor in filtered_monitors:
        for metric in metrics:
            if metric["name"] in monitor["query"]:
                if not args.unused:
                    print(monitor["name"], "\n")
                    print(monitor["url"], "\n")
                    print(monitor["query"], "\n")
                    print("---------\n")

                used_monitor_metrics.add(metric["name"])
            else:
                unused_monitor_metrics.add(metric["name"])

    # avoid scenario where one monitor does reference the metric
    # but a latter monitor DOES NOT reference it. when that happens
    # we want to ensure we remove the metric name so it doesn't
    # accidentally get marked later as being unused.
    for metric in used_monitor_metrics:
        try:
            unused_monitor_metrics.remove(metric)
        except KeyError:
            pass

    if not args.unused:
        format_title("unused metrics in monitors")

        for m in unused_monitor_metrics:
            print(m)

    format_title("unused metrics in monitors and dashboards")

    unused_metrics = unused_dashboard_metrics.intersection(
        unused_monitor_metrics)

    for m in unused_metrics:
        print(m)

    print(f"\n{len(unused_metrics)} out of {len(metrics)} metrics are unused.")


if __name__ == '__main__':
    process()
```

So the first thing to notice is that we accept various flags to control the behaviour of the script:

- `-p/--pattern`: provide a regex pattern and the script will filter the number of dashboards/monitors to those whose 'title' matches the given pattern.
- `-m/--metrics`: we need a set of metrics to search for, these should be provided in CSV format.
- `-u/--unused`: indicate we only care about seeing what metrics are unused and can be safely deleted.

When it comes to providing the list of metrics to this script I originally started by writing a bash script to parse the various service code we had. This was fraught with errors and generally was very brittle. I then realized that it would be better to use the Datadog API to tell me what metrics our services were producing!

There were two problems: the first was that there was only one API endpoint I could really use for this and it only returned metrics that were reported within the last 24hrs. 

This might be an issue for you, but we were using Datadog's various UI based tools to help us identify 'big hitters' as far as our expensive custom metrics were concerned and these big hitters were producing lots of metrics on an hourly basis (so 24hrs was an acceptable caveat for our use case).

The second issue we had was that the API endpoint wasn't supported for Python, only Curl (and possibly Ruby too?). 

My kingdom for API consistency!

![aaaah](../../images/aaaah.webp)

This meant I needed a way to combine Python with some bash scripting. So let's start by looking at how we would call the Python script:

```
time python3 searcher.py -m $(...)
```

Everything within the subprocess `$(...)` will be me using the Curl API endpoint. There's a lot of `grep`, `sed` and other unix utilities at play, and I appreciate some of y'all probably could have done better with `awk` but I just struggle to get along with `awk` most of the time and so I tend to reach for other more commonly understood unix tools. 

**Shield your eyes...**

```
export api_key=foo app_key=bar && \
curl -s -H "DD-API-KEY: ${api_key}" -H "DD-APPLICATION-KEY: ${app_key}" \
  "https://api.datadoghq.com/api/v1/search?q=metrics:YOUR_METRIC_NAMESPACE" | \
  jq -r '.results.metrics' | \
  egrep '"' | \
  sed -e 's/  //' | sed 's/"//g' | sed 's/,//' | \
  grep r'^YOUR_METRIC_NAMESPACE\.' | \
  gsed r's/\.\(avg\|count\|median\|95percentile\|max\)$//' | \
  sort | uniq | tr "\n" "," | \
  sed 's/, /,/g'
```

Ultimately it takes the JSON response from the API endpoint and coerces it into CSV that's then passed into the `--metrics` flag.

There are two other key sections of the Python script:

1. `process()`
2. `find_graphs()`

The `process()` function is the coordinator. It controls spinning up a threadpool for calling the Datadog API (which isn't asynchronous) while spinning up a process pool for handling the CPU heavy parts of the tasks such as recursively traversing the dashboard graphs.

> **Note**: my logic for tracking metrics was a lot simpler originally, but executing code within separate process pools (which unlike a threadpool weren't sharing memory) introduced their own unique challenges as far as data isolation was concerned, so things got a bit tricksy in places.

I then use a `set` abstract data type to help identify the unused metrics by way of a intersection operation.

The `find_graphs()` function also became a bit more complex once I introduced a process pool because I needed to pass in extra contextual data (e.g. dashboard title/url) that the function itself didn't require, but the main/parent process did need as part of the subprocess response/output.

Also the logic for locating the 'query' for a specific graph (e.g. where the metric is going to be referenced) worked fine for my needs, but it likely is not catching every possible case. I just kept adding conditions to the logic until I had avoided all errors being raised from the near 2000 different dashboards/monitors (which felt like "good enough" to me).

I also won't explain that piece of the code, cause yer know  
**...there be dragons**.

![there be dragons](../../images/there-be-dragons.gif)

## `DISTRIBUTION` metric type?

OK, so earlier I mentioned that we were able to utilize the script output to help us identify places where we might be able to switch from a TIMER and/or HISTOGRAM metric to a DISTRIBUTION metric. 

### Metric Types

Overall there are five distinct metric 'types' to be aware of ([docs](https://docs.datadoghq.com/developers/metrics/types/?tab=histogram#metric-types)).

Two of those metric types require additional clarification with regards to the cost implications associated with instrumenting your code to report metric data.

- `HISTOGRAM`
- `DISTRIBUTION`

> **Note**: there is also a `TIMER` metric type, which is a subset of `HISTOGRAM` ([docs](https://docs.datadoghq.com/developers/metrics/dogstatsd_metrics_submission/?tab=python#timer)) so for the purpose of this post we'll consider them the same and just discuss `HISTOGRAM`.

The key differences between `HISTOGRAM` and `DISTRIBUTION` are...

| | `HISTOGRAM` | `DISTRIBUTION` |
| --- | --- | --- |
| **Multiple Metrics** | YES | YES |
| **Percentile Aggregations** | PARTIAL | YES |
| **Tag Filtering** | NO | YES |

### Multiple Metrics

The `HISTOGRAM` metric type will generate five unique custom metrics to represent different aggregations: 

For example, if you report a histogram metric `foo.bar`, then this will result in the following metrics being created (representing different aggregation types):

- `foo.bar.avg`
- `foo.bar.count`
- `foo.bar.median`
- `foo.bar.max`
- `foo.bar.95percentile`

The `DISTRIBUTION` metric type will generate one metric, but provide multiple aggregations via the Datadog UI. 

The aggregations for a `DISTRIBUTION` metric type are:

- `max`
- `min`
- `avg`
- `sum`
- `count` 

Now although the `DISTRIBUTION` aggregations may well be applied to a 'single' metric, _internally_ Datadog considers each aggregation a _unique_ metric. This means at face value a `DISTRIBUTION` metric type is no more cost effective than a `HISTOGRAM` with regards to the calculation of 'custom metrics' (as they both ultimately yield five individual metrics). 

But this isn't necessarily the case, as a `DISTRIBUTION` metric type has the added ability to [filter tags](#tag-filtering) thus reducing the number of calculated custom metrics (see below for details).

### Percentile Aggregations

The `HISTOGRAM` metric type provides a `95percentile`, while the `DISTRIBUTION` metric type _can_ provide a `p95` along with `p50`, `p75`, `p90` and `p99` but these aggregations need to be manually generated via the Datadog UI.

Each percentile aggregation for the `DISTRIBUTION` metric type is internally considered a _unique_ metric and thus is subject to Datadog's custom metric cost implications.

### Tag Filtering

The `DISTRIBUTION` metric type allows tags to be filtered, thus reducing the potential number of custom metrics Datadog will charge us for. This is not possible with any other metric type.

## Choosing `HISTOGRAM` or `DISTRIBUTION` ?

The fact that the `DISTRIBUTION` metric type enables tag filtering is an important consideration when choosing between it and a `HISTOGRAM`. 

As an example, we provide our services an abstraction over Datadog's client library in a shared 'metrics' package. We provide a `timer()` abstraction that enables teams to measure the time it takes for their code to run. This timer abstraction uses the `DISTRIBUTION` metric type by default (see implementation below).

```
import asyncio
import functools
import logging
import time

import datadog

logger = logging.getLogger(__name__)


def serialize_tags(tags):
    """serialize a dict of tags into a list of colon separated strings.
    this is the format for the datadog client's key-value tags.
    """
    return ['{}:{}'.format(k, v) for k, v in tags.items() if v is not None]


class Timer():
    def __init__(self, metric_name, distribution, tags=None):
        self.metric_name = metric_name
        self.distribution = distribution
        self.tags = tags

    def __call__(self, func):
        """decorator implementation."""

        if asyncio.iscoroutinefunction(func):
            @functools.wraps(func)
            async def wrap_timer(*args, **kwargs):
                start_time = time.perf_counter() * 1000
                result = await func(*args, **kwargs)
                end_time = time.perf_counter() * 1000

                run_time = end_time - start_time
                self.distribution(self.metric_name, run_time, self.tags)

                return result
        else:
            @functools.wraps(func)
            def wrap_timer(*args, **kwargs):
                start_time = time.perf_counter() * 1000
                result = func(*args, **kwargs)
                end_time = time.perf_counter() * 1000

                run_time = end_time - start_time
                self.distribution(self.metric_name, run_time, self.tags)

                return result

        return wrap_timer

    def __enter__(self):
        self.start_time = time.perf_counter() * 1000
        return self

    def __exit__(self, *args):
        end_time = time.perf_counter() * 1000
        run_time = end_time - self.start_time

        self.distribution(self.metric_name, run_time, self.tags)


class Metrics(datadog.DogStatsd):
    """Statsd client with a better tagging interface.

    Supports passing tags as a list of colon separated strings (this is the
    Datadog client's expected format), while also suporting tags passed as a
    dictionary.

    If write_metrics is False, metrics will only be logged.

    Usage:
        # tags as a dictionary...
            metrics = bf_metrics.Metrics(
                namespace='foo',
                host='localhost',
                constant_tags={'foo': 'bar'},
            )
            metrics.incr('foo', tags={'baz': 'qux'})
        # tags as a list...
            metrics = bf_metrics.Metrics(
                namespace='foo',
                host='localhost',
                constant_tags=['foo:bar'},
            )
            metrics.incr('foo', tags=['baz:qux'])

            metrics.timer("foo")
            def slow_operation():
                ...
    """

    def __init__(self, *args, **kwargs):
        self.write_metrics = kwargs.pop('write_metrics', True)
        self.incr = self.increment
        self.decr = self.decrement

        constant_tags = kwargs.pop('constant_tags', {})
        if isinstance(constant_tags, dict):
            constant_tags = serialize_tags(constant_tags)

        super(Metrics, self).__init__(constant_tags=constant_tags, *args, **kwargs)

    def _report(self, metric, metric_type, value, tags, sample_rate):
        if isinstance(tags, dict):
            tags = serialize_tags(tags)

        all_tags = tags.extend(self.constant_tags) if tags else self.constant_tags

        msg = "metric {metric}:{value} tags={tags}"
        logger.debug(msg.format(msg, metric=metric, value=value, tags=all_tags))

        if not self.write_metrics:
            return

        super(Metrics, self)._report(
            metric, metric_type, value,
            tags=tags, sample_rate=sample_rate)

    def timer(self, metric_name, tags=None):
        """calculate distributed run time of workload.
        supported as both a decorator and context manager.
        """
        return Timer(metric_name, self.distribution, tags)
```

By using a `DISTRIBUTION` metric type it means that we're able to reduce the number of custom metrics while also allowing consumers to opt-in to percentile aggregations if they require them, and to again utilize tag filtering to help constrain the number of custom metrics.

The `DISTRIBUTION` and `HISTOGRAM` have overlapping aggregations (`count`, `avg`, `max`) which means if you do not require an aggregation outside of those specific ones, then choosing a `DISTRIBUTION` metric type would be better as you can utilize tag filtering to help reduce the number of custom metrics.

If you do require a percentile aggregation then the trade-off you need to make is between whether a `HISTOGRAM` (with `95percentile` available by default) is more cost effective than a `DISTRIBUTION` which with percentiles added will add up to more individual metrics but with tag filtering available might still end up being more cost effective overall as you can't filter your high-cardinality tags with a `HISTOGRAM`/`TIMER`.

### The `DISTRIBUTION` percentile 'custom metric' difference

The way Datadog calculates the number of 'custom metrics' is slightly different (and more costly) for percentile aggregations of a `DISTRIBUTION` metric type. **It got me like...**

![NO God NO](../../images/no-god-no.webp)

Let's just recap what a 'custom metric' is...

> A custom metric is uniquely identified by a combination of a metric name and tag values (including the host tag). — [Datadog](https://docs.datadoghq.com/account_management/billing/custom_metrics/)

Constrast this with the `DISTRIBUTION` percentiles, which take into account every _potentially_ queryable varation of a metric. 

Imagine your metric has three tags A, B and C. Datadog calculates the number of custom metrics like so:

```
- each tag value of {A} 
- each tag value of {B}
- each tag value of {C}
- each tag value of {A,B}
- each tag value of {A,C}
- each tag value of {B,C}
- each tag value of {A,B,C}
- {*}
```

Datadog's rationale for this difference is...

> The reason we have to store percentiles for each potentially queryable tag value combination is to preserve mathematical accuracy of your values; unlike the non-percentile aggregations which can be mathematically accurate when reaggregated (i.e the global max is the maximum of the maxes), you can't calculate the globally accurate p95 by recombining p95s.
