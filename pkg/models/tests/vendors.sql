DROP TABLE IF EXISTS vendors;

CREATE TABLE vendors (
	id serial4 NOT NULL,
	mine_id int4 NOT NULL,
	username text NOT NULL,
	created_at text NULL,
	updated_at text NULL,
	"type" text NULL,
	login text NULL,
	"name" text NULL,
	company text NULL,
	blog text NULL,
	"location" text NULL,
	email text NULL,
	twitter_username text NULL,
	repos int4 NULL,
	gists int4 NULL,
	followers int4 NULL,
	"following" int4 NULL,
	location_id int4 NOT NULL,
	CONSTRAINT vendors_mine_id_username_key UNIQUE (mine_id, username),
	CONSTRAINT vendors_pkey PRIMARY KEY (id)
);
INSERT INTO vendors
(id, mine_id, username, created_at, updated_at, "type", login, "name", company, blog, "location", email, twitter_username, repos, gists, followers, "following", location_id)
VALUES(1405399, 5, 'scanoss-qg', '2021-01-26T11:34:42Z', '2024-07-12T17:26:23Z', 'User', 'scanoss-qg', NULL, NULL, NULL, '', '', NULL, 18, 0, 2, 2, 30539);
INSERT INTO vendors
(id, mine_id, username, created_at, updated_at, "type", login, "name", company, blog, "location", email, twitter_username, repos, gists, followers, "following", location_id)
VALUES(1405465, 5, 'mscasso-scanoss', '2020-12-06T22:46:26Z', '2024-07-13T11:58:47Z', 'User', 'mscasso-scanoss', 'Mariano Scasso', NULL, NULL, '', '', NULL, 14, 0, 3, 2, 30539);
INSERT INTO vendors
(id, mine_id, username, created_at, updated_at, "type", login, "name", company, blog, "location", email, twitter_username, repos, gists, followers, "following", location_id)
VALUES(1320559, 5, 'lijie', '2009-01-06T14:50:28Z', '2024-07-19T16:55:53Z', 'User', 'lijie', 'Li Jie', 'Tencent', NULL, 'Shenzhen, China', '', NULL, 20, 0, 35, 9, 5528);
INSERT INTO vendors
(id, mine_id, username, created_at, updated_at, "type", login, "name", company, blog, "location", email, twitter_username, repos, gists, followers, "following", location_id)
VALUES(2108508, 5, 'sumitbsn', '2015-03-04T15:16:20Z', '2023-03-29T10:47:37Z', 'User', 'sumitbsn', 'sumit kumar', 'coder guy', 'https://sumitbsn.pythonanywhere.com', 'Bangalore', 'sumitbsn@gmail.com', NULL, 37, 0, 11, 8, 15609);
INSERT INTO vendors
(id, mine_id, username, created_at, updated_at, "type", login, "name", company, blog, "location", email, twitter_username, repos, gists, followers, "following", location_id)
VALUES(1485164, 5, 'ChrisADR', '2015-09-23T17:40:52Z', '2023-03-10T18:23:29Z', 'User', 'ChrisADR', 'Christopher Díaz', '@gentoo @CodeLabora  ', 'https://blogs.gentoo.org/chrisadr/', 'Lima, PE', 'chrisadr@gentoo.org', NULL, 22, 0, 32, 0, 16642);
INSERT INTO vendors
(id, mine_id, username, created_at, updated_at, "type", login, "name", company, blog, "location", email, twitter_username, repos, gists, followers, "following", location_id)
VALUES(1406024, 5, 'perezale', '2012-08-10T23:46:19Z', '2024-07-12T11:20:45Z', 'User', 'perezale', 'Alejandro Pérez', NULL, NULL, 'Tandil', '', NULL, 115, 3, 14, 25, 29682);
INSERT INTO vendors
(id, mine_id, username, created_at, updated_at, "type", login, "name", company, blog, "location", email, twitter_username, repos, gists, followers, "following", location_id)
VALUES(1405464, 5, 'scanossjs', '2019-05-22T12:58:13Z', '2024-07-02T11:48:18Z', 'User', 'scanossjs', 'Juan M Salamanca', '@scanoss ', 'https://www.scanoss.com', '', '', NULL, 10, 0, 0, 0, 30539);
INSERT INTO vendors
(id, mine_id, username, created_at, updated_at, "type", login, "name", company, blog, "location", email, twitter_username, repos, gists, followers, "following", location_id)
VALUES(3236055, 5, 'jeronimoortiz', '2021-05-07T18:03:51Z', '2024-02-21T05:43:31Z', 'User', 'jeronimoortiz', 'Jeronimo-Ortiz.lr', NULL, 'https://www.linkedin.com/in/jeronimo-ortiz-2b71041b2/', 'Buenos Aires, Argentina', '', NULL, 2, 0, 0, 0, 39105);
INSERT INTO vendors
(id, mine_id, username, created_at, updated_at, "type", login, "name", company, blog, "location", email, twitter_username, repos, gists, followers, "following", location_id)
VALUES(3277960, 5, 'scanossjeronimo', '2024-04-08T12:39:43Z', '2024-06-27T17:43:05Z', 'User', 'scanossjeronimo', 'Jeronimo Ortiz', 'ScanOSS', NULL, 'Argentina', '', NULL, 7, 0, 1, 1, 14761);
INSERT INTO vendors
(id, mine_id, username, created_at, updated_at, "type", login, "name", company, blog, "location", email, twitter_username, repos, gists, followers, "following", location_id)
VALUES(1197052, 5, 'ulm', '2010-02-18T21:31:26Z', '2024-04-15T12:27:19Z', 'User', 'ulm', 'Ulrich Müller', NULL, NULL, 'Germany', 'ulm@gentoo.org', NULL, 49, 1, 29, 0, 12103);
INSERT INTO vendors
(id, mine_id, username, created_at, updated_at, "type", login, "name", company, blog, "location", email, twitter_username, repos, gists, followers, "following", location_id)
VALUES(351225, 5, 'torvalds', '2011-09-03T15:26:22Z', '2023-11-12T20:08:30Z', 'User', 'torvalds', 'Linus Torvalds', 'Linux Foundation', NULL, 'Portland, OR', '', NULL, 31, 0, 212023, 0, 23624);
INSERT INTO vendors
(id, mine_id, username, created_at, updated_at, "type", login, "name", company, blog, "location", email, twitter_username, repos, gists, followers, "following", location_id)
VALUES(1620783, 5, 'TorstenScheck', '2015-01-14T14:16:23Z', '2020-03-26T09:32:43Z', 'User', 'TorstenScheck', NULL, NULL, NULL, '', '', NULL, 4, 0, 6, 0, 30539);
