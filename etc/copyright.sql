create database copyright character set utf8;
use copyright

drop table if exists vote;
drop table if exists account_content;
drop table if exists aution;
drop table if exists account;
drop table if exists content;
create table account
(
   account_id           int not null primary key auto_increment,
   email                 varchar(50),
   username             varchar(30),
   identity_id          varchar(100),
   address              varchar(256)
);
CREATE UNIQUE INDEX account_email_uindex ON copyright2.account (email);
CREATE UNIQUE INDEX account_name_uindex ON copyright2.account (username);
alter table account comment '账户表';


create table content
(
   content_id           int not null primary key auto_increment,
   title                varchar(100),
   content              blob,
   content_hash         varchar(256),
   ts                   timestamp
);

create table account_content
(
   account_id           int,
   content_id           int,
   content_hash         varchar(256),
   percent              int,
   price                int,
   sell_price           int,
   sell_percent         int,
   ts                   timestamp,
   status               varchar(2),
   tokenid              int
);

alter table account_content add constraint FK_Reference_2 foreign key (account_id)
      references account (account_id) on delete restrict on update restrict;

alter table account_content add constraint FK_Reference_3 foreign key (content_id)
      references content (content_id) on delete restrict on update restrict;

create table aution
(
   content_hash         varchar(256),
   account_id           int,
   percent              int,
   price                int,
   ts                   timestamp,
   end_ts               timestamp default now()
);

alter table aution add constraint FK_Reference_4 foreign key (account_id)
      references account (account_id) on delete restrict on update restrict;

create table vote
(
   account_id           int,
   vote_id              int primary key auto_increment,
   content_hash         varchar(256),
   vote_time            timestamp,
   comment              varchar(100)
);

alter table vote comment '投票表，一个账户一个图片，只能投一票，一票代表50pxc';

alter table vote add constraint FK_Reference_5 foreign key (account_id)
      references account (account_id) on delete restrict on update restrict;

delete from vote;
delete from aution;
delete from account_content;
delete from content;
delete from account;