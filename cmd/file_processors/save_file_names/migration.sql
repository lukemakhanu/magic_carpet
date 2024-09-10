show create table leagues;
show create table live_scores_files;
show create table ls_files;
show create table match_date;
show create table match_odds;
show create table matches;
show create table o_files;
show create table odds_file;
show create table processed_files;
show create table scheduled_time;
show create table season_weeks;
show create table seasons;
show create table sn_wks;
show create table sns;
show create table ssn_weeks;
show create table ssns;
show create table teams;
show create table tvts;
show create table tvts_old;
show create table tvtss;
show create table winning_outcome_files;
show create table wo_files;


CREATE TABLE `leagues` (
  `league_id` smallint(4) NOT NULL AUTO_INCREMENT,
  `client_id` int(11) NOT NULL,
  `league` varchar(100) NOT NULL,
  `league_abbrv` varchar(50) NOT NULL,
  `created` datetime NOT NULL,
  `modified` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`league_id`),
  KEY `client_id` (`client_id`),
  KEY `league` (`league`),
  KEY `league_abbrv` (`league_abbrv`),
  KEY `created` (`created`)
);

CREATE TABLE `live_scores_files` (
  `live_score_file_id` bigint(20) NOT NULL AUTO_INCREMENT,
  `ls_file_name` varchar(300) NOT NULL,
  `ls_dir` varchar(300) NOT NULL,
  `country` varchar(100) NOT NULL,
  `ext_id` bigint(30) NOT NULL,
  `project_id` int(6) NOT NULL,
  `competition_id` int(6) NOT NULL,
  `created` datetime NOT NULL,
  `modified` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`live_score_file_id`),
  UNIQUE KEY `country_2` (`country`,`ext_id`),
  KEY `ls_file_name` (`ls_file_name`),
  KEY `ls_dir` (`ls_dir`),
  KEY `country` (`country`),
  KEY `ext_id` (`ext_id`),
  KEY `project_id` (`project_id`),
  KEY `competition_id` (`competition_id`),
  KEY `created` (`created`),
  KEY `modified` (`modified`)
);

CREATE TABLE `ls_files` (
  `ls_file_id` bigint(20) NOT NULL AUTO_INCREMENT,
  `ls_file_name` varchar(300) NOT NULL,
  `ls_dir` varchar(300) NOT NULL,
  `country` varchar(100) NOT NULL,
  `ls_ext_id` bigint(20) NOT NULL,
  `project_id` int(6) NOT NULL,
  `competition_id` int(6) NOT NULL,
  `created` datetime NOT NULL,
  `modified` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`ls_file_id`),
  UNIQUE KEY `ls_file_name_2` (`ls_file_name`,`project_id`),
  KEY `ls_file_name` (`ls_file_name`),
  KEY `ls_dir` (`ls_dir`),
  KEY `country` (`country`),
  KEY `ls_ext_id` (`ls_ext_id`),
  KEY `created` (`created`),
  KEY `modified` (`modified`),
  KEY `project_id` (`project_id`),
  KEY `competition_id` (`competition_id`)
);

CREATE TABLE `match_date` (
  `match_date_id` int(11) NOT NULL AUTO_INCREMENT,
  `league_id` int(11) NOT NULL,
  `client_id` int(11) NOT NULL,
  `date` date NOT NULL,
  `created` datetime NOT NULL,
  `modified` datetime NOT NULL,
  PRIMARY KEY (`match_date_id`),
  UNIQUE KEY `league_id_2` (`league_id`,`client_id`,`date`),
  KEY `league_id` (`league_id`),
  KEY `date` (`date`),
  KEY `created` (`created`),
  KEY `modified` (`modified`),
  KEY `client_id` (`client_id`)
);

CREATE TABLE `match_odds` (
  `match_odds_id` bigint(30) NOT NULL AUTO_INCREMENT,
  `round_id` bigint(20) NOT NULL,
  `parent_id` bigint(20) NOT NULL,
  `country` varchar(20) NOT NULL,
  `created` datetime NOT NULL,
  `modified` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`match_odds_id`),
  UNIQUE KEY `round_id_2` (`round_id`,`parent_id`,`country`),
  KEY `round_id` (`round_id`),
  KEY `parent_id` (`parent_id`),
  KEY `country` (`country`),
  KEY `created` (`created`),
  KEY `modified` (`modified`)
);

CREATE TABLE `matches` (
  `match_id` bigint(20) NOT NULL AUTO_INCREMENT,
  `season_week_id` int(11) NOT NULL,
  `home_team_id` smallint(3) NOT NULL,
  `away_team_id` smallint(3) NOT NULL,
  `status` enum('inactive','active','cancelled','finished') NOT NULL,
  `created` datetime NOT NULL,
  `modified` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`match_id`),
  KEY `season_week_id` (`season_week_id`),
  KEY `home_team_id` (`home_team_id`),
  KEY `away_team_id` (`away_team_id`),
  KEY `status` (`status`),
  KEY `created` (`created`),
  KEY `modified` (`modified`)
);

CREATE TABLE `o_files` (
  `odds_file_id` bigint(20) NOT NULL AUTO_INCREMENT,
  `odds_file_name` varchar(300) NOT NULL,
  `file_directory` varchar(300) NOT NULL,
  `country` varchar(50) NOT NULL,
  `parent_id` varchar(30) NOT NULL,
  `competition_id` varchar(100) NOT NULL,
  `match_id` int(11) NOT NULL DEFAULT '0',
  `created` datetime NOT NULL,
  `modified` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`odds_file_id`),
  UNIQUE KEY `country_2` (`country`,`parent_id`),
  KEY `odds_file_name` (`odds_file_name`),
  KEY `file_directory` (`file_directory`),
  KEY `country` (`country`),
  KEY `created` (`created`),
  KEY `modified` (`modified`),
  KEY `parent_id` (`parent_id`),
  KEY `competition_id` (`competition_id`),
  KEY `match_id` (`match_id`)
);

CREATE TABLE `odds_file` (
  `odds_file_id` bigint(20) NOT NULL AUTO_INCREMENT,
  `odds_file_name` varchar(300) NOT NULL,
  `file_directory` varchar(300) NOT NULL,
  `country` varchar(50) NOT NULL,
  `parent_id` varchar(30) NOT NULL,
  `competition_id` varchar(100) NOT NULL,
  `match_id` int(11) NOT NULL DEFAULT '0',
  `created` datetime NOT NULL,
  `modified` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`odds_file_id`),
  KEY `odds_file_name` (`odds_file_name`),
  KEY `file_directory` (`file_directory`),
  KEY `country` (`country`),
  KEY `created` (`created`),
  KEY `modified` (`modified`),
  KEY `parent_id` (`parent_id`),
  KEY `competition_id` (`competition_id`),
  KEY `match_id` (`match_id`)
);

CREATE TABLE `processed_files` (
  `processed_file_id` bigint(30) NOT NULL AUTO_INCREMENT,
  `odds_file_id` bigint(20) NOT NULL,
  `date_processed` datetime NOT NULL,
  `client_id` int(11) NOT NULL,
  `live_score_found` enum('pending','yes','no') NOT NULL,
  `live_score_dir` varchar(300) NOT NULL,
  `winning_outcome_found` enum('pending','yes','no','') NOT NULL,
  `file_status` enum('pending','processed','failed') NOT NULL,
  `season_week_id` int(11) NOT NULL DEFAULT '0',
  `created` datetime NOT NULL,
  `modified` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `winning_outcome_dir` varchar(300) NOT NULL,
  PRIMARY KEY (`processed_file_id`),
  UNIQUE KEY `odds_file_id_2` (`odds_file_id`,`client_id`,`season_week_id`),
  KEY `odds_file_id` (`odds_file_id`),
  KEY `date_processed` (`date_processed`),
  KEY `client_id` (`client_id`),
  KEY `live_score_found` (`live_score_found`),
  KEY `live_score_dir` (`live_score_dir`),
  KEY `winning_outcome_found` (`winning_outcome_found`),
  KEY `created` (`created`),
  KEY `modified` (`modified`),
  KEY `winning_outcome_dir` (`winning_outcome_dir`),
  KEY `season_week_id` (`season_week_id`),
  KEY `file_status` (`file_status`)
);

CREATE TABLE `scheduled_time` (
  `scheduled_time_id` int(11) NOT NULL AUTO_INCREMENT,
  `scheduled_time` varchar(100) NOT NULL,
  `competition_id` smallint(4) NOT NULL,
  `status` enum('active','inactive') NOT NULL,
  `created` datetime NOT NULL,
  `modified` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`scheduled_time_id`),
  UNIQUE KEY `scheduled_time_2` (`scheduled_time`,`competition_id`),
  KEY `scheduled_time` (`scheduled_time`),
  KEY `competition_id` (`competition_id`),
  KEY `status` (`status`),
  KEY `created` (`created`),
  KEY `modified` (`modified`)
);

CREATE TABLE `season_weeks` (
  `season_week_id` int(11) NOT NULL AUTO_INCREMENT,
  `season_id` int(11) NOT NULL,
  `week_number` smallint(3) NOT NULL,
  `status` enum('inactive','active','cancelled','finished') NOT NULL,
  `start_time` datetime NOT NULL,
  `end_time` datetime NOT NULL,
  `created` datetime NOT NULL,
  `modified` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`season_week_id`),
  UNIQUE KEY `season_id_2` (`season_id`,`week_number`,`start_time`),
  KEY `season_id` (`season_id`),
  KEY `week_number` (`week_number`),
  KEY `start_time` (`start_time`),
  KEY `end_time` (`end_time`),
  KEY `created` (`created`),
  KEY `modified` (`modified`),
  KEY `status` (`status`)
);

CREATE TABLE `seasons` (
  `season_id` int(11) NOT NULL AUTO_INCREMENT,
  `league_id` int(11) NOT NULL,
  `created` datetime NOT NULL,
  `modified` datetime NOT NULL,
  PRIMARY KEY (`season_id`),
  UNIQUE KEY `match_date_id` (`league_id`),
  KEY `project_round_id` (`league_id`),
  KEY `start_time` (`created`),
  KEY `end_time` (`modified`)
);

CREATE TABLE `sn_wks` (
  `season_week_id` int(11) NOT NULL AUTO_INCREMENT,
  `league_id` smallint(4) NOT NULL,
  `season_id` int(11) NOT NULL,
  `week_number` smallint(3) NOT NULL,
  `status` enum('inactive','active','cancelled','finished') NOT NULL,
  `start_time` datetime NOT NULL,
  `end_time` datetime NOT NULL,
  `created` datetime NOT NULL,
  `modified` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`season_week_id`),
  UNIQUE KEY `season_id_3` (`season_id`,`week_number`,`start_time`),
  KEY `season_id` (`season_id`),
  KEY `week_number` (`week_number`),
  KEY `start_time` (`start_time`),
  KEY `end_time` (`end_time`),
  KEY `created` (`created`),
  KEY `modified` (`modified`),
  KEY `status` (`status`),
  KEY `league_id` (`league_id`)
);

CREATE TABLE `sns` (
  `season_id` int(11) NOT NULL AUTO_INCREMENT,
  `league_id` smallint(4) NOT NULL,
  `status` enum('active','inactive','cancelled') NOT NULL,
  `created` datetime NOT NULL,
  `modified` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`season_id`),
  KEY `league_id` (`league_id`),
  KEY `status` (`status`),
  KEY `created` (`created`),
  KEY `modified` (`modified`)
);

CREATE TABLE `ssn_weeks` (
  `season_week_id` int(11) NOT NULL AUTO_INCREMENT,
  `league_id` smallint(4) NOT NULL,
  `season_id` int(11) NOT NULL,
  `week_number` smallint(3) NOT NULL,
  `status` enum('inactive','active','cancelled','finished') NOT NULL,
  `start_time` datetime NOT NULL,
  `end_time` datetime NOT NULL,
  `created` datetime NOT NULL,
  `modified` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`season_week_id`),
  UNIQUE KEY `season_id_3` (`season_id`,`week_number`,`start_time`),
  KEY `season_id` (`season_id`),
  KEY `week_number` (`week_number`),
  KEY `start_time` (`start_time`),
  KEY `end_time` (`end_time`),
  KEY `created` (`created`),
  KEY `modified` (`modified`),
  KEY `status` (`status`),
  KEY `league_id` (`league_id`)
);

CREATE TABLE `ssns` (
  `season_id` int(11) NOT NULL AUTO_INCREMENT,
  `league_id` smallint(4) NOT NULL,
  `status` enum('active','inactive','cancelled') NOT NULL,
  `created` datetime NOT NULL,
  `modified` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`season_id`),
  KEY `league_id` (`league_id`),
  KEY `status` (`status`),
  KEY `created` (`created`),
  KEY `modified` (`modified`)
);

CREATE TABLE `teams` (
  `team_id` int(11) NOT NULL AUTO_INCREMENT,
  `league_id` smallint(4) NOT NULL,
  `team_name` varchar(100) NOT NULL,
  `team_abbrv` varchar(20) NOT NULL,
  `created` datetime NOT NULL,
  `modified` datetime NOT NULL,
  PRIMARY KEY (`team_id`),
  KEY `league_id` (`league_id`),
  KEY `team_name` (`team_name`),
  KEY `team_abbrv` (`team_abbrv`),
  KEY `created` (`created`),
  KEY `modified` (`modified`)
);

CREATE TABLE `tvts` (
  `tvt_id` int(11) NOT NULL AUTO_INCREMENT,
  `client_id` smallint(11) NOT NULL,
  `league_id` int(4) NOT NULL,
  `week_number` smallint(4) NOT NULL,
  `game_number` smallint(4) NOT NULL,
  `teams` varchar(20) NOT NULL,
  `season_count` smallint(4) NOT NULL,
  `created` datetime NOT NULL,
  `modified` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`tvt_id`),
  KEY `client_id` (`client_id`),
  KEY `league_id` (`league_id`),
  KEY `week_number` (`week_number`),
  KEY `game_number` (`game_number`),
  KEY `season_count` (`season_count`),
  KEY `created` (`created`),
  KEY `modified` (`modified`),
  KEY `teams` (`teams`)
);

CREATE TABLE `tvts_old` (
  `tvt_id` int(11) NOT NULL AUTO_INCREMENT,
  `client_id` smallint(11) NOT NULL,
  `league_id` int(4) NOT NULL,
  `week_number` smallint(4) NOT NULL,
  `game_number` smallint(4) NOT NULL,
  `teams` varchar(20) NOT NULL,
  `season_count` smallint(4) NOT NULL,
  `created` datetime NOT NULL,
  `modified` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`tvt_id`),
  KEY `client_id` (`client_id`),
  KEY `league_id` (`league_id`),
  KEY `week_number` (`week_number`),
  KEY `game_number` (`game_number`),
  KEY `season_count` (`season_count`),
  KEY `created` (`created`),
  KEY `modified` (`modified`),
  KEY `teams` (`teams`)
);

CREATE TABLE `tvtss` (
  `tvt_id` int(11) NOT NULL AUTO_INCREMENT,
  `client_id` smallint(11) NOT NULL,
  `league_id` int(4) NOT NULL,
  `week_number` smallint(4) NOT NULL,
  `game_number` smallint(4) NOT NULL,
  `teams` varchar(20) NOT NULL,
  `season_count` smallint(4) NOT NULL,
  `created` datetime NOT NULL,
  `modified` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`tvt_id`),
  KEY `client_id` (`client_id`),
  KEY `league_id` (`league_id`),
  KEY `week_number` (`week_number`),
  KEY `game_number` (`game_number`),
  KEY `season_count` (`season_count`),
  KEY `created` (`created`),
  KEY `modified` (`modified`),
  KEY `teams` (`teams`)
);

CREATE TABLE `winning_outcome_files` (
  `wo_file_id` bigint(20) NOT NULL AUTO_INCREMENT,
  `wo_file_name` varchar(300) NOT NULL,
  `wo_dir` varchar(300) NOT NULL,
  `country` varchar(100) NOT NULL,
  `ext_id` bigint(20) NOT NULL,
  `project_id` int(6) NOT NULL,
  `competition_id` int(6) NOT NULL,
  `created` datetime NOT NULL,
  `modified` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`wo_file_id`),
  UNIQUE KEY `ext_id_3` (`ext_id`,`country`),
  KEY `wo_file_name` (`wo_file_name`),
  KEY `wo_dir` (`wo_dir`),
  KEY `country` (`country`),
  KEY `ext_id` (`ext_id`),
  KEY `project_id` (`project_id`),
  KEY `competition_id` (`competition_id`),
  KEY `created` (`created`),
  KEY `modified` (`modified`)
);

CREATE TABLE `wo_files` (
  `wo_file_id` bigint(20) NOT NULL AUTO_INCREMENT,
  `wo_file_name` varchar(300) NOT NULL,
  `wo_dir` varchar(300) NOT NULL,
  `country` varchar(100) NOT NULL,
  `wo_ext_id` bigint(20) NOT NULL,
  `project_id` int(6) NOT NULL,
  `competition_id` int(6) NOT NULL,
  `created` datetime NOT NULL,
  `modified` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`wo_file_id`),
  UNIQUE KEY `wo_file_name_2` (`wo_file_name`,`project_id`),
  KEY `wo_file_name` (`wo_file_name`),
  KEY `wo_dir` (`wo_dir`),
  KEY `country` (`country`),
  KEY `wo_ext_id` (`wo_ext_id`),
  KEY `project_id` (`project_id`),
  KEY `competition_id` (`competition_id`),
  KEY `created` (`created`),
  KEY `modified` (`modified`)
);