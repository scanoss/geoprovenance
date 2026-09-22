DROP TABLE IF EXISTS github_contributors;

CREATE TABLE github_contributors (
	purl_name text NOT NULL,
	contributor text NOT NULL,
	CONSTRAINT github_contributors_pkey PRIMARY KEY (purl_name,contributor)
);
INSERT INTO github_contributors
(purl_name, contributor)
VALUES('torvalds/uemacs', 'torvalds');
INSERT INTO github_contributors
(purl_name, contributor)
VALUES('torvalds/uemacs', 'tfarina');
INSERT INTO github_contributors
(purl_name, contributor)
VALUES('torvalds/uemacs', 'penberg');
INSERT INTO github_contributors
(purl_name, contributor)
VALUES('torvalds/uemacs', 'vonbrand');
INSERT INTO github_contributors
(purl_name, contributor)
VALUES('torvalds/uemacs', 'lijie');
INSERT INTO github_contributors
(purl_name, contributor)
VALUES('torvalds/uemacs', 'naota');
INSERT INTO github_contributors
(purl_name, contributor)
VALUES('torvalds/uemacs', 'thiagofarina');
INSERT INTO github_contributors
(purl_name, contributor)
VALUES('torvalds/uemacs', 'ulm');
INSERT INTO github_contributors
(purl_name, contributor)
VALUES('scanoss/engine', 'jeronimoortiz');
INSERT INTO github_contributors
(purl_name, contributor)
VALUES('scanoss/engine', 'scanossjeronimo');
INSERT INTO github_contributors
(purl_name, contributor)
VALUES('scanoss/engine', 'scanoss');
INSERT INTO github_contributors
(purl_name, contributor)
VALUES('scanoss/engine', 'mscasso-scanoss');
INSERT INTO github_contributors
(purl_name, contributor)
VALUES('scanoss/engine', 'scanoss-cs');
INSERT INTO github_contributors
(purl_name, contributor)
VALUES('scanoss/engine', 'scanoss-qg');
INSERT INTO github_contributors
(purl_name, contributor)
VALUES('scanoss/minr', 'scanoss-qg');
INSERT INTO github_contributors
(purl_name, contributor)
VALUES('scanoss/engine', 'scanossjs');
INSERT INTO github_contributors
(purl_name, contributor)
VALUES('scanoss/engine', 'mengzhuo');
INSERT INTO github_contributors
(purl_name, contributor)
VALUES('scanoss/engine', 'perezale');
INSERT INTO github_contributors
(purl_name, contributor)
VALUES('scanoss/engine', 'eeisegn');
INSERT INTO github_contributors
(purl_name, contributor)
VALUES('scanoss/engine', 'jens-erdmann');
INSERT INTO github_contributors
(purl_name, contributor)
VALUES('scanoss/engine', 'superkaiy');
INSERT INTO github_contributors
(purl_name, contributor)
VALUES('scanoss/engine', 'vpenso');


DROP TABLE IF EXISTS too_many_contributors;

CREATE TABLE
  too_many_contributors (purl_name text NOT NULL, mine_id integer NOT NULL);
INSERT INTO too_many_contributors (purl_name, mine_id) VALUES ('torvalds/linux', 5);
