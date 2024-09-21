CREATE TABLE `match_requests` (
  `match_request_id` bigint(40) NOT NULL,
  `player_id` bigint(30) NOT NULL,
  `start_time` datetime NOT NULL,
  `end_time` datetime NOT NULL,
  `early_finish` enum('no','yes') NOT NULL DEFAULT 'no',
  `played` enum('no','yes') NOT NULL DEFAULT 'no',
  `created` datetime DEFAULT NULL,
  `modified` datetime NOT NULL DEFAULT current_timestamp()
);

ALTER TABLE `match_requests`
  ADD KEY `match_request_id` (`match_request_id`),
  ADD KEY `player_id` (`player_id`),
  ADD KEY `start_time` (`start_time`),
  ADD KEY `end_time` (`end_time`),
  ADD KEY `early_finish` (`early_finish`),
  ADD KEY `created` (`created`),
  ADD KEY `played` (`played`),
  ADD KEY `modified` (`modified`);

ALTER TABLE `match_requests`
  MODIFY `match_request_id` bigint(40) NOT NULL AUTO_INCREMENT;

CREATE TABLE `players` (
  `prayer_id` bigint(30) NOT NULL,
  `profile_tag` varchar(400) NOT NULL,
  `status` enum('active','inactive','suspended','blocked') NOT NULL,
  `created` datetime NOT NULL,
  `modified` datetime NOT NULL DEFAULT current_timestamp()
);

ALTER TABLE `players`
  ADD PRIMARY KEY (`prayer_id`),
  ADD UNIQUE KEY `profile_tag` (`profile_tag`);

ALTER TABLE `players`
  MODIFY `prayer_id` bigint(30) NOT NULL AUTO_INCREMENT;

CREATE TABLE `selected_matches` (
  `selected_matches_id` bigint(40) NOT NULL,
  `player_id` bigint(30) NOT NULL,
  `match_request_id` bigint(40) NOT NULL,
  `parent_match_id` varchar(50) NOT NULL,
  `created` datetime NOT NULL,
  `modified` datetime NOT NULL
);

ALTER TABLE `selected_matches`
  ADD PRIMARY KEY (`selected_matches_id`),
  ADD UNIQUE KEY `match_request_id_2` (`match_request_id`,`parent_match_id`),
  ADD UNIQUE KEY `player_id_2` (`player_id`,`parent_match_id`),
  ADD KEY `match_request_id` (`match_request_id`),
  ADD KEY `parent_match_id` (`parent_match_id`),
  ADD KEY `created` (`created`),
  ADD KEY `modified` (`modified`),
  ADD KEY `player_id` (`player_id`);

ALTER TABLE `selected_matches`
  MODIFY `selected_matches_id` bigint(40) NOT NULL AUTO_INCREMENT;

ALTER TABLE `players` CHANGE `prayer_id` `player_id` BIGINT(30) NOT NULL AUTO_INCREMENT;
