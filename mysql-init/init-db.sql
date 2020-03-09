USE pathwalker

DROP TABLE IF EXISTS paths;
CREATE TABLE paths (
  `path_id`           CHAR(36),
  `path_description`  VARCHAR(36),
  `survey_ids`        VARCHAR(255),
  PRIMARY KEY(`path_id`)
);

DROP TABLE IF EXISTS surveys;
CREATE TABLE surveys (
  `survey_id`           CHAR(36),
  `path_id`             CHAR(36),
  `survey_date`         DATE,
  `survey_submitted_by` VARCHAR(255),
  `survey_detail`       TEXT,
  `image_ids`           VARCHAR(255),
  PRIMARY KEY(`survey_id`)
);

DROP TABLE IF EXISTS images;
CREATE TABLE images (
  `image_id`           CHAR(36),
  `path_id`            CHAR(36),
  `filename`           VARCHAR(255),
  `image_description`  TEXT,
  `image_latitude`     FLOAT,
  `image_longitude`    FLOAT,
  PRIMARY KEY(`image_id`)
);

INSERT INTO paths
  (`path_id`,`path_description`,`survey_ids`)
  VALUES
    ('HAN/3', 'Hanmer to Bronington', 'test-survey-001'),
    ('HAN/17', 'Hanmer to Ellesmere', '');

INSERT INTO surveys
  (`survey_id`, `path_id`, `survey_date`, `survey_submitted_by`, `survey_detail`, `image_ids`)
  VALUES
    ('test-survey-001', 'HAN/3', '2020-03-12', 'mbeaney641@gmail.com', 'oooh, this is a lovely path.  such scenery!  such beauty!  pity about the litter', '');
