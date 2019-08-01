--
-- PostgreSQL database dump
--

-- Dumped from database version 11.1
-- Dumped by pg_dump version 11.1

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: categories; Type: TABLE; Schema: public; Owner: sdduncan
--

CREATE TABLE public.categories (
    club_id integer NOT NULL,
    category text NOT NULL
);


ALTER TABLE public.categories OWNER TO sdduncan;

--
-- Name: clubs; Type: TABLE; Schema: public; Owner: sdduncan
--

CREATE TABLE public.clubs (
    club_id integer NOT NULL,
    name text NOT NULL,
    bio text,
    show_members boolean NOT NULL
);


ALTER TABLE public.clubs OWNER TO sdduncan;

--
-- Name: followers; Type: TABLE; Schema: public; Owner: sdduncan
--

CREATE TABLE public.followers (
    net_id text NOT NULL,
    club_id integer NOT NULL,
    admin boolean,
    member boolean,
    name text NOT NULL,
    club_name text NOT NULL,
    id text NOT NULL
);


ALTER TABLE public.followers OWNER TO sdduncan;

--
-- Name: links; Type: TABLE; Schema: public; Owner: sdduncan
--

CREATE TABLE public.links (
    club_id integer NOT NULL,
    link_type text NOT NULL,
    descriptor text NOT NULL
);


ALTER TABLE public.links OWNER TO sdduncan;

--
-- Name: postings; Type: TABLE; Schema: public; Owner: sdduncan
--

CREATE TABLE public.postings (
    posting_id integer NOT NULL,
    club_id integer NOT NULL,
    title text NOT NULL,
    member_post boolean NOT NULL,
    blurb text,
    long_blurb text,
    creation_time timestamp without time zone
);


ALTER TABLE public.postings OWNER TO sdduncan;

--
-- Name: test; Type: TABLE; Schema: public; Owner: sdduncan
--

CREATE TABLE public.test (
    club_id text NOT NULL,
    name text NOT NULL,
    bio text,
    show_members text NOT NULL
);


ALTER TABLE public.test OWNER TO sdduncan;

--
-- Name: users; Type: TABLE; Schema: public; Owner: sdduncan
--

CREATE TABLE public.users (
    net_id text NOT NULL,
    name text NOT NULL,
    bio text NOT NULL
);


ALTER TABLE public.users OWNER TO sdduncan;

--
-- Data for Name: categories; Type: TABLE DATA; Schema: public; Owner: sdduncan
--

COPY public.categories (club_id, category) FROM stdin;
1	sport
2	sport
3	sport
1	land
2	water
3	land
\.


--
-- Data for Name: clubs; Type: TABLE DATA; Schema: public; Owner: sdduncan
--

COPY public.clubs (club_id, name, bio, show_members) FROM stdin;
1	Fencing	We like fences	t
2	Water Polo	We like water	t
3	Wrestling	We like to wrestle	t
\.


--
-- Data for Name: followers; Type: TABLE DATA; Schema: public; Owner: sdduncan
--

COPY public.followers (net_id, club_id, admin, member, name, club_name, id) FROM stdin;
sdduncan	3	f	t	Sean	Wrestling	sdduncan3
\.


--
-- Data for Name: links; Type: TABLE DATA; Schema: public; Owner: sdduncan
--

COPY public.links (club_id, link_type, descriptor) FROM stdin;
2	Facebook	facebook.com
2	Instagram	instagram.com
\.


--
-- Data for Name: postings; Type: TABLE DATA; Schema: public; Owner: sdduncan
--

COPY public.postings (posting_id, club_id, title, member_post, blurb, long_blurb, creation_time) FROM stdin;
1	2	Welcome New Members	f	Welcome!	Welcome to the team!	2018-11-11 14:44:40.633545
2	2	Game Today	f	Game today!	Game vs Harvard at 4 pm today!	2018-11-11 14:44:40.651344
3	3	Fencing Match	t	come!	come to the match today!	2018-11-13 14:44:40.131414
4	3	Fencing Pregame	f	come!	yeeeeeeeet!	2018-11-13 14:44:40.131414
5	2	Joint Pregame	t	come!	yeeeeeeeet!	2018-11-13 14:44:40.131414
\.


--
-- Data for Name: test; Type: TABLE DATA; Schema: public; Owner: sdduncan
--

COPY public.test (club_id, name, bio, show_members) FROM stdin;
1	jan	ahhh	True
2	asdf	ahhh	yeet
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: sdduncan
--

COPY public.users (net_id, name, bio) FROM stdin;
1	bbbb	af
sdduncan	Sean	F yea
popovici	Michael	Does this work
	sean	haha suck it
kyleem	Kyle	the mac macmillan
ssh2	Sadie	hi
\.


--
-- Name: clubs clubs_name_key; Type: CONSTRAINT; Schema: public; Owner: sdduncan
--

ALTER TABLE ONLY public.clubs
    ADD CONSTRAINT clubs_name_key UNIQUE (name);


--
-- Name: clubs clubs_pkey; Type: CONSTRAINT; Schema: public; Owner: sdduncan
--

ALTER TABLE ONLY public.clubs
    ADD CONSTRAINT clubs_pkey PRIMARY KEY (club_id);


--
-- Name: followers followers_pkey; Type: CONSTRAINT; Schema: public; Owner: sdduncan
--

ALTER TABLE ONLY public.followers
    ADD CONSTRAINT followers_pkey PRIMARY KEY (id);


--
-- Name: postings postings_pkey; Type: CONSTRAINT; Schema: public; Owner: sdduncan
--

ALTER TABLE ONLY public.postings
    ADD CONSTRAINT postings_pkey PRIMARY KEY (posting_id);


--
-- Name: test test_name_key; Type: CONSTRAINT; Schema: public; Owner: sdduncan
--

ALTER TABLE ONLY public.test
    ADD CONSTRAINT test_name_key UNIQUE (name);


--
-- Name: test test_pkey; Type: CONSTRAINT; Schema: public; Owner: sdduncan
--

ALTER TABLE ONLY public.test
    ADD CONSTRAINT test_pkey PRIMARY KEY (club_id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: sdduncan
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (net_id);


--
-- PostgreSQL database dump complete
--

