--
-- PostgreSQL database dump
--

-- Dumped from database version 14.5 (Ubuntu 14.5-1.pgdg22.04+1)
-- Dumped by pg_dump version 14.5 (Ubuntu 14.5-1.pgdg22.04+1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: url_mappings; Type: TABLE; Schema: public; Owner: ya2k
--

CREATE TABLE public.url_mappings (
    id bigint NOT NULL,
    tinyurl character varying(7) NOT NULL,
    longurl character varying(1024) NOT NULL
);


ALTER TABLE public.url_mappings OWNER TO ya2k;

--
-- Name: url_mappings_id_seq; Type: SEQUENCE; Schema: public; Owner: ya2k
--

CREATE SEQUENCE public.url_mappings_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.url_mappings_id_seq OWNER TO ya2k;

--
-- Name: url_mappings_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: ya2k
--

ALTER SEQUENCE public.url_mappings_id_seq OWNED BY public.url_mappings.id;


--
-- Name: url_mappings id; Type: DEFAULT; Schema: public; Owner: ya2k
--

ALTER TABLE ONLY public.url_mappings ALTER COLUMN id SET DEFAULT nextval('public.url_mappings_id_seq'::regclass);


--
-- Data for Name: url_mappings; Type: TABLE DATA; Schema: public; Owner: ya2k
--

COPY public.url_mappings (id, tinyurl, longurl) FROM stdin;
\.


--
-- Name: url_mappings_id_seq; Type: SEQUENCE SET; Schema: public; Owner: ya2k
--

SELECT pg_catalog.setval('public.url_mappings_id_seq', 177004, true);


--
-- Name: url_mappings_longurl_idx; Type: INDEX; Schema: public; Owner: ya2k
--

CREATE UNIQUE INDEX url_mappings_longurl_idx ON public.url_mappings USING btree (longurl);


--
-- Name: url_mappings_tinyurl_idx; Type: INDEX; Schema: public; Owner: ya2k
--

CREATE UNIQUE INDEX url_mappings_tinyurl_idx ON public.url_mappings USING btree (tinyurl);


--
-- PostgreSQL database dump complete
--
