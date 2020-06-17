# outfit7-expertise-test

## Design decisions
  1. Programming language: Go
  2. Http framework: chi
  3. Virtualization: Yes, docker(-compose)
  4. Networking: Using traefik and my personal domain (local.verbic.pro) for local routing and development. User requires to remember no ports.
  5. Storage: Decided to add Redis as an ad hoc storage to allow for horizontal scaling as well as a vertical one.

## Questions and assumptions about the task:

### Data
1. What format is the data delivered in? <br/>
   Assumption: data is returned in a json format. <br/>
   Assumption #2: data to be served is always latest. <br/>
   Assumption #3: data does not contain duplicates or malformed data. <br/>
   Possibilities: <br/>
    - Data consists of a json array of possibly unsorted ad networks, differentiate between networks by using object fields.
      Example:
      ```json
      {
        "data":[
          {
            "name":"Facebook",
            "country":"si",
            "type":"banner",
            "score":10.0
          },
          {
            "name":"Facebook",
            "country":"si",
            "type":"interstitial",
            "score":10.0
          },
          {
            "name":"Facebook",
            "country":"si",
            "type":"video",
            "score":10.0
          },
          {
            "name":"Facebook",
            "country":"it",
            "type":"banner",
            "score":10.0
          },
          {
            "name":"AdMob",
            "country":"si",
            "type":"banner",
            "score":10.0
          }
        ]
      }
      ```

    - Data is already filtered by country and type.
      Example:
      ```json
      {
        "data":[
          {
            "country":"si",
            "banner":[
              {
                "name":"Facebook",
                "score":8.0
              },
              {
                "name":"AdMob",
                "score":3.0
              },
              {
                "name":"Huawei Ads",
                "score":5.0
              }
            ],
            "video":[
              {
                "name":"Facebook",
                "score":10.0
              },
              {
                "name":"AdMob",
                "score":9.9
              },
              {
                "name":"Huawei Ads",
                "score":2.1
              }
            ]
          },
          {
            "country":"cr",
            "banner":[
              {
                "name":"Facebook",
                "score":5.0
              },
              {
                "name":"AdMob",
                "score":10.0
              },
              {
                "name":"Huawei Ads",
                "score":8.0
              }
            ],
            "interstitial":[
              {
                "name":"Facebook",
                "score":10.0
              },
              {
                "name":"AdMob",
                "score":9.9
              },
              {
                "name":"Huawei Ads",
                "score":2.1
              }
            ]
          }
        ]
      }
      ```
2. What other piece of data might be useful.
   - OS
   - Device type (tablet, smartphone, smart tv, ...)

### Storage
1. Should data be continuously collected directly from the pipe's output?
   To ensure the minimum downtime in case of system failure in other services, collection should only happen once per processing stage lifecycle.
2. Should data be stored in memory/database?
   - Database storage would create possibility for long term monitoring.
   - In memory storage is set up quickly and works when only the latest dataset has to be kept. (Assumption #2)
3. Should data be stored post transformation?
   - Since the only transformation to happen is a change in order and filtering, storing the data would be useful for long term monitoring and testing.

### Architecture
1. What security measures need to be implemented on API level?
   - Specifications show no required security settings, authorization should still be implemented.
2. Transformation at API call or separate concurrent engine with in memory storage?
   - Given the design choice that a single call to pipeline should be implemented, a simple engine to transform the data could be implemented.
   - Only filtering through the data should be done at api call.
3. What happens if no relevant data exists?
   - Return a random AdNetwork.
   - Could possibly group countries by continent and return a country in the same continent.
4. REST or GraphQL?
   - Specifications call for REST, GraphQL could speed up client side data parsing.
5. Possible filters at request time and how to implement?
   - MVP product only requires country so no special filtering systems are required to be implemented
   - Could possibly add a filter by device type and/or operating system.
6. Exceptions?
   - Create a config file which describes possible exceptions.
   - Should this be handled at api call or pipe data parsing?
7. Testing?
   - Api testing
   - Unit tests
8. Mocking the pipeline.
   - MVP could possibly only require a reset to original value as a replacement for the processing pipelines.
   - Should be implemented in such a way that reload could also be called from /update endpoint.



### System schema


```
Collection Service ─┬ Warmup ┬ Sort ─ Exceptions ─ Storage
                    └ Reload ┘                       │
                        ↑                            │
API ────────────────┬ Update                         │
                    └ List ──── Filter ──────────────┘
```
