# magic_carpet
Implements file processors, data manipulators and apis for instant virtuals

# file_processors/save_file_names
This app reads directories with odds,winning outcomes and livescores and saves this file names into mysql database for further prcessing.

# file_processors/save_files_in_redis
This app saves files names into redis sets for further processing. The primary set that we send data to is NEW_RAW_WINNING_OUTCOME.

# file_processors/save_keys
This app reads from NEW_RAW_WINNING_OUTCOME and from there we figure out odds and livescores. This keys the final keys we will use for parent_match_id for winning outcomes, live scores and odds.