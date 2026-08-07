-- ============================================================================
-- DSS (Data Source Service) Database Initialization Script
-- ============================================================================
-- This script creates all required schemas and tables for the DSS service.
-- The users table in activedirectoryusers schema is created dynamically
-- by the application based on HR metadata fields provided via API.
-- ============================================================================

-- ============================================================================
-- SCHEMA: activedirectoryusers
-- Purpose: Stores HR/user data and configuration
-- ============================================================================
CREATE SCHEMA IF NOT EXISTS activedirectoryusers;

-- DSS Configuration table (single-row configuration)
CREATE TABLE IF NOT EXISTS activedirectoryusers.dss_config (
    id INT PRIMARY KEY DEFAULT 1,
    customer_before_hr_sync_refactoring BOOLEAN NOT NULL DEFAULT 0
);

-- Insert default configuration (update if exists, insert if not)
INSERT INTO activedirectoryusers.dss_config (id, customer_before_hr_sync_refactoring) VALUES (1, 0)
ON DUPLICATE KEY UPDATE customer_before_hr_sync_refactoring = 0;

-- HR Mapping table (stores HR field metadata)
CREATE TABLE IF NOT EXISTS activedirectoryusers.hr_mapping (
    id INT AUTO_INCREMENT PRIMARY KEY,
    source_column VARCHAR(255) NOT NULL,
    destination_column VARCHAR(255) NOT NULL,
    column_description VARCHAR(255) NOT NULL,
    ui_representation VARCHAR(255) NULL,
    optional BIT(1) NOT NULL DEFAULT b'0',
    part_of_shield_id BIT(1) NOT NULL DEFAULT b'0'
);

-- Note: The 'users' table is created dynamically by the application
-- via the HR data API endpoint. It will have columns based on
-- the HR metadata fields provided. Example structure:
-- CREATE TABLE activedirectoryusers.users (
--     ShieldId VARCHAR(255) PRIMARY KEY,
--     FirstName VARCHAR(255),
--     LastName VARCHAR(255),
--     Email VARCHAR(255),
--     ... (other HR fields)
--     isDeleted BIT(1) NOT NULL,
--     isUpdated BIT(1) NOT NULL,
--     CONSTRAINT unique_shield_id UNIQUE (...)
-- );

-- ============================================================================
-- SCHEMA: shieldsitedb
-- Purpose: Stores site-specific metadata and user information
-- ============================================================================
CREATE SCHEMA IF NOT EXISTS shieldsitedb;

-- View Metadata table
CREATE TABLE IF NOT EXISTS shieldsitedb.view_metadata (
    id                  BIGINT AUTO_INCREMENT PRIMARY KEY,
    platform_name       VARCHAR(128)     NOT NULL,
    section_name        VARCHAR(128)     NOT NULL,
    display_name        VARCHAR(128)     NOT NULL,
    display_name_nested VARCHAR(128)     NULL,
    display_order       INT              NOT NULL,
    db_field            VARCHAR(128)     NOT NULL,
    db_field2           VARCHAR(128)     NULL,
    db_field_nested     VARCHAR(128)     NULL,
    db_field_type       INT              NOT NULL,
    is_conversation     BIT              NOT NULL,
    is_displayed        BIT              NOT NULL,
    is_unique_id        BIT              DEFAULT b'0' NOT NULL
);

-- Custom Filter Metadata table
CREATE TABLE IF NOT EXISTS shieldsitedb.custom_filter_metadata (
    filter_id                           BIGINT AUTO_INCREMENT PRIMARY KEY,
    datasource_name                     VARCHAR(255)                                                           NULL,
    db_field_name                       VARCHAR(255)                                                           NULL,
    display_name                        VARCHAR(255)                                                           NULL,
    favorite                            BIT                                                                    NOT NULL,
    content_type                        VARCHAR(255)                                                           NULL,
    filter_section                      ENUM ('ECOM', 'PARTICIPANT', 'OTHER', 'PRODUCT') DEFAULT 'PARTICIPANT' NULL,
    filterType                          INT                                                                    NULL,
    is_auto_complete                    BIT                                                                    NULL,
    participant_autocomplete_field_name VARCHAR(255)                                                           NULL,
    predefinedValues                    VARCHAR(255)                                                           NULL,
    is_displayed                        BIT                                                                    DEFAULT b'0' NULL
);

-- Participant View Metadata table
CREATE TABLE IF NOT EXISTS shieldsitedb.participant_view_metadata (
    participant_view_metadata_id BIGINT AUTO_INCREMENT PRIMARY KEY,
    category_name                VARCHAR(255) NULL,
    custom_filters_favorite      BIT          NULL,
    display_name                 VARCHAR(255) NOT NULL,
    display_order                INT          NULL,
    elastic_field                VARCHAR(255) NULL,
    Participant_favorite         BIT          NULL,
    type                         VARCHAR(255) NOT NULL
);

-- User Type table
CREATE TABLE IF NOT EXISTS shieldsitedb.user_type_tbl (
    user_type_id BIGINT AUTO_INCREMENT PRIMARY KEY,
    name         VARCHAR(255) NULL,
    CONSTRAINT name UNIQUE (name)
);

-- Group table
CREATE TABLE IF NOT EXISTS shieldsitedb.group_tbl (
    group_id              BIGINT AUTO_INCREMENT PRIMARY KEY,
    group_type            VARCHAR(255) NOT NULL,
    is_ad_synced          BIT          NOT NULL,
    group_name            VARCHAR(255) NULL,
    organizationunit_name VARCHAR(255) NULL,
    group_unique_name     VARCHAR(255) NULL,
    CONSTRAINT group_unique_name UNIQUE (group_unique_name)
);

-- Monitored Users table
CREATE TABLE IF NOT EXISTS shieldsitedb.monitored_users_tbl (
    monitored_user_id BIGINT AUTO_INCREMENT PRIMARY KEY,
    date_updated      DATE                                      NULL,
    firstName         VARCHAR(255)                              NULL,
    is_active         BIT                                       DEFAULT b'1' NULL,
    lastName          VARCHAR(255)                              NULL,
    username          VARCHAR(255)                              NULL,
    group_id          BIGINT                                    NULL,
    user_type_id      BIGINT                                    NULL,
    shield_id         VARCHAR(255)                              DEFAULT '' NULL COMMENT 'Shield ID',
    time_updated      DATETIME                                  DEFAULT CURRENT_TIMESTAMP NULL,
    CONSTRAINT unique_shield_id UNIQUE (shield_id),
    CONSTRAINT fk_user_type FOREIGN KEY (user_type_id) REFERENCES user_type_tbl (user_type_id),
    CONSTRAINT fk_group FOREIGN KEY (group_id) REFERENCES group_tbl (group_id)
);

-- ============================================================================
-- SCHEMA: shieldcoredb
-- Purpose: Core database for task groups, plugins, MDEs, and policies
-- ============================================================================
CREATE SCHEMA IF NOT EXISTS shieldcoredb;

-- Deployment Base Role table
CREATE TABLE IF NOT EXISTS shieldcoredb.deployment_base_role (
    Id   INT AUTO_INCREMENT PRIMARY KEY,
    Name VARCHAR(45) NOT NULL
);

-- Deployment Role table
CREATE TABLE IF NOT EXISTS shieldcoredb.deployment_role (
    Id            INT AUTO_INCREMENT PRIMARY KEY,
    Name          VARCHAR(128)  NOT NULL,
    LocalFilename VARCHAR(4096) NOT NULL,
    BaseRoleId    INT           NOT NULL,
    CONSTRAINT FK_BaseRole_Role FOREIGN KEY (BaseRoleId) REFERENCES deployment_base_role (Id) ON UPDATE CASCADE ON DELETE CASCADE
);

-- Note: Index is automatically created by the foreign key constraint above, no need for explicit CREATE INDEX

-- DR JSON Template table
CREATE TABLE IF NOT EXISTS shieldcoredb.dr_json_template (
    Id              INT AUTO_INCREMENT PRIMARY KEY,
    Name            VARCHAR(128)                       NOT NULL,
    JsonTemplateStr MEDIUMTEXT                         NOT NULL,
    DateAdded       DATETIME                           DEFAULT CURRENT_TIMESTAMP NULL
);

-- EML JSON Template table
CREATE TABLE IF NOT EXISTS shieldcoredb.eml_json_template (
    Id              INT AUTO_INCREMENT PRIMARY KEY,
    Name            VARCHAR(128) NOT NULL,
    JsonTemplateStr TEXT         NOT NULL,
    DateAdded       DATETIME     NULL
);

-- Participants HTML Template table
CREATE TABLE IF NOT EXISTS shieldcoredb.participants_html_template (
    Id              INT AUTO_INCREMENT PRIMARY KEY,
    Name            VARCHAR(128) NOT NULL,
    HtmlTemplateStr MEDIUMTEXT   NOT NULL
);

-- MDE Plugin table
CREATE TABLE IF NOT EXISTS shieldcoredb.mde_plugin (
    Id                INT AUTO_INCREMENT PRIMARY KEY,
    Name              VARCHAR(128)                       NOT NULL,
    Path              VARCHAR(1024)                      NOT NULL,
    MainClassFullName VARCHAR(1024)                      NOT NULL,
    IsDynamic         BIT                                NOT NULL,
    DateAdded         DATETIME                           DEFAULT CURRENT_TIMESTAMP NOT NULL,
    ParameterTemplate VARCHAR(8192)                       NOT NULL
);

-- MDE Group table
CREATE TABLE IF NOT EXISTS shieldcoredb.mde_group (
    Id   INT AUTO_INCREMENT PRIMARY KEY,
    Name VARCHAR(128) NOT NULL
);

-- MDE Group Template table
CREATE TABLE IF NOT EXISTS shieldcoredb.mde_group_template (
    Id   INT AUTO_INCREMENT PRIMARY KEY,
    Name VARCHAR(128) NOT NULL
);

-- MDE table
CREATE TABLE IF NOT EXISTS shieldcoredb.mde (
    Id          INT AUTO_INCREMENT PRIMARY KEY,
    Name        VARCHAR(128)  NOT NULL,
    StepNumber  INT           NOT NULL,
    CommandLine VARCHAR(4096) NOT NULL,
    MdeGroupId  INT           NOT NULL,
    MdePluginId INT           NOT NULL,
    LogLevelId  INT           NOT NULL,
    IsActive    BIT           NOT NULL,
    CONSTRAINT FK_MdeGroup_Mde FOREIGN KEY (MdeGroupId) REFERENCES mde_group (Id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT FK_MdePlugin_Mde FOREIGN KEY (MdePluginId) REFERENCES mde_plugin (Id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE INDEX FK_LogLevel_Mde_idx ON shieldcoredb.mde (LogLevelId);
CREATE INDEX FK_MdeGroup_Mde_idx ON shieldcoredb.mde (MdeGroupId);
CREATE INDEX FK_MdePlugin_Mde_idx ON shieldcoredb.mde (MdePluginId);

-- MDE Template table
CREATE TABLE IF NOT EXISTS shieldcoredb.mde_template (
    Id                 INT AUTO_INCREMENT PRIMARY KEY,
    Name               VARCHAR(128)  NOT NULL,
    StepNumber         INT           NOT NULL,
    CommandLine        VARCHAR(4096) NOT NULL,
    MdeGroupTemplateId INT           NOT NULL,
    MdePluginId        INT           NOT NULL,
    LogLevelId         INT           NOT NULL,
    IsActive           BIT           NOT NULL,
    CONSTRAINT FK_MdeGroupTemplate_MdeTemplate FOREIGN KEY (MdeGroupTemplateId) REFERENCES mde_group_template (Id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT FK_MdePlugin_MdeTemplate FOREIGN KEY (MdePluginId) REFERENCES mde_plugin (Id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE INDEX FK_LogLevel_MdeTemplate_idx ON shieldcoredb.mde_template (LogLevelId);
CREATE INDEX FK_MdeGroupTemplate_MdeTemplate_idx ON shieldcoredb.mde_template (MdeGroupTemplateId);
CREATE INDEX FK_MdePlugin_MdeTemplate_idx ON shieldcoredb.mde_template (MdePluginId);

-- Plugin table
CREATE TABLE IF NOT EXISTS shieldcoredb.plugin (
    Id                INT AUTO_INCREMENT PRIMARY KEY,
    Name              VARCHAR(128)                            NOT NULL,
    TypeId            INT                                     NOT NULL,
    SubTypeId         INT                                     NOT NULL,
    Path              VARCHAR(1024)                           NOT NULL,
    MainClassFullName VARCHAR(1024)                           NOT NULL,
    IsDynamic         BIT                                     NOT NULL,
    DateAdded         DATETIME                                 DEFAULT CURRENT_TIMESTAMP NOT NULL,
    ParameterTemplate VARCHAR(8192)                            DEFAULT '' NOT NULL,
    TaskGroupTypeId   INT                                     NOT NULL
);

CREATE INDEX FK_PluginSubType_Plugin_idx ON shieldcoredb.plugin (SubTypeId);
CREATE INDEX FK_PluginType_Plugin_idx ON shieldcoredb.plugin (TypeId);

-- Policy Status table
CREATE TABLE IF NOT EXISTS shieldcoredb.policy_status (
    Id   INT AUTO_INCREMENT PRIMARY KEY,
    Name VARCHAR(255) NOT NULL UNIQUE
);

-- Policy Step table
CREATE TABLE IF NOT EXISTS shieldcoredb.policy_step (
    Id   INT AUTO_INCREMENT PRIMARY KEY,
    Name VARCHAR(255) NOT NULL UNIQUE
);

-- Policy Type table
CREATE TABLE IF NOT EXISTS shieldcoredb.policy_type (
    Id   INT AUTO_INCREMENT PRIMARY KEY,
    Name VARCHAR(256) NOT NULL
);

-- Policy table
CREATE TABLE IF NOT EXISTS shieldcoredb.policy (
    Id              INT          NOT NULL PRIMARY KEY,
    Name            VARCHAR(256) NOT NULL,
    Description     MEDIUMTEXT   NULL,
    Reason          VARCHAR(256) NULL,
    PolicyType      INT          NOT NULL,
    Status          INT          NULL,
    Step            INT          NOT NULL,
    IsEnabled       BIT          NULL,
    CreatedByUserId INT          NOT NULL,
    CreatedDate     DATETIME     NOT NULL,
    AuthorizedDate  DATETIME     NULL,
    CONSTRAINT policy_ibfk_1 FOREIGN KEY (PolicyType) REFERENCES policy_type (Id),
    CONSTRAINT policy_ibfk_2 FOREIGN KEY (Status) REFERENCES policy_status (Id),
    CONSTRAINT policy_ibfk_3 FOREIGN KEY (Step) REFERENCES policy_step (Id)
);

CREATE INDEX PolicyType ON shieldcoredb.policy (PolicyType);
CREATE INDEX Status ON shieldcoredb.policy (Status);
CREATE INDEX Step ON shieldcoredb.policy (Step);

-- Policy Base Retention table
CREATE TABLE IF NOT EXISTS shieldcoredb.policy_base_retention (
    Id                  INT NOT NULL PRIMARY KEY,
    TaskGroupId         INT NULL,
    RetentionPeriod     INT NULL,
    RetentionPeriodType INT NULL,
    CountryId           INT NULL,
    ArchiveId           INT NULL,
    CONSTRAINT policy_base_retention_ibfk_1 FOREIGN KEY (Id) REFERENCES policy (Id) ON UPDATE CASCADE ON DELETE CASCADE
    -- Note: Foreign key constraint for TaskGroupId is added later after task_group table is created
);

CREATE INDEX policy_base_retention_retention_period_type_Id_fk ON shieldcoredb.policy_base_retention (RetentionPeriodType);
CREATE INDEX policy_base_retention_task_group_Id_fk ON shieldcoredb.policy_base_retention (TaskGroupId);

-- Schedule Type table
CREATE TABLE IF NOT EXISTS shieldcoredb.schedule_type (
    Id   INT AUTO_INCREMENT PRIMARY KEY,
    Name VARCHAR(128) NOT NULL
);

-- Sequence Store table (for ID generation)
CREATE TABLE IF NOT EXISTS shieldcoredb.seq_store (
    SeqName  VARCHAR(255) NOT NULL PRIMARY KEY,
    SeqValue BIGINT       NOT NULL
);

-- Insert initial sequence value for Policy IDs
INSERT IGNORE INTO shieldcoredb.seq_store (SeqName, SeqValue)
VALUES ('POLICY.ID.PK', 101);

-- Task Group Type table
CREATE TABLE IF NOT EXISTS shieldcoredb.task_group_type (
    Id   INT       NOT NULL PRIMARY KEY,
    Name VARCHAR(45) NOT NULL
);

-- Task Subgroup table
CREATE TABLE IF NOT EXISTS shieldcoredb.task_subgroup (
    Id   INT AUTO_INCREMENT PRIMARY KEY,
    Name VARCHAR(128) NOT NULL
);

-- Task Type table
CREATE TABLE IF NOT EXISTS shieldcoredb.task_type (
    Id   INT AUTO_INCREMENT PRIMARY KEY,
    Name VARCHAR(128) NOT NULL
);

-- Task Group Template table
CREATE TABLE IF NOT EXISTS shieldcoredb.task_group_template (
    Id                                     INT AUTO_INCREMENT PRIMARY KEY,
    Name                                   VARCHAR(128)     NOT NULL,
    ScheduleTypeId                         INT              NOT NULL,
    StartTime                              DATETIME         NOT NULL,
    TimeInterval                           BIGINT           NOT NULL,
    IsRunImmediatelyAfterTaskGroupFinished BIT              NOT NULL,
    MdeGroupTemplateId                     INT              NOT NULL,
    IsMessageQueue                         BIT              NOT NULL,
    IsCompleteMissingTaskInRecovery        BIT              DEFAULT b'0' NOT NULL,
    PlatformName                           VARCHAR(128)     NOT NULL,
    EcommTypeId                            INT              NOT NULL,
    JsonMapperId                           INT              NOT NULL,
    RetentionDays                          INT              NOT NULL,
    IsCopyToArchive                        BIT              NOT NULL,
    RelatedArchiveConnectionId             INT              DEFAULT 0 NOT NULL,
    IsActive                               BIT              NOT NULL,
    DefaultLogLevelId                      INT              NOT NULL,
    EmlMapperId                            INT              DEFAULT 0 NULL,
    HtmlMapperId                           INT              NULL,
    IsHistoric                             BIT              DEFAULT b'0' NOT NULL,
    IsArchive                              BIT              DEFAULT b'1' NOT NULL,
    IsArchiveRetention                     BIT              DEFAULT b'1' NOT NULL,
    IsMigrationTaskGroup                   BIT              DEFAULT b'0' NOT NULL,
    CONSTRAINT FK_DrJsonTemplate_TaskGroupTemplate FOREIGN KEY (JsonMapperId) REFERENCES dr_json_template (Id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT FK_MdeGroupTemplate_TaskGroupTemplate FOREIGN KEY (MdeGroupTemplateId) REFERENCES mde_group_template (Id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT FK_ScheduleType_TaskGroupTemplate FOREIGN KEY (ScheduleTypeId) REFERENCES schedule_type (Id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT task_group_template_ibfk_1 FOREIGN KEY (EmlMapperId) REFERENCES eml_json_template (Id),
    CONSTRAINT task_group_template_ibfk_2 FOREIGN KEY (HtmlMapperId) REFERENCES participants_html_template (Id)
);

CREATE INDEX FK_DefaultLogLevel_TaskGroupTemplate_idx ON shieldcoredb.task_group_template (DefaultLogLevelId);
CREATE INDEX FK_DrJsonTemplate_TaskGroupTemplate_idx ON shieldcoredb.task_group_template (JsonMapperId);
CREATE INDEX FK_EcommType_TaskGroupTemplate_idx ON shieldcoredb.task_group_template (EcommTypeId);
CREATE INDEX FK_EmlJsonTemplate_TaskGroupTemplate ON shieldcoredb.task_group_template (EmlMapperId);
CREATE INDEX FK_MdeGroupTemplate_TaskGroupTemplate_idx ON shieldcoredb.task_group_template (MdeGroupTemplateId);
CREATE INDEX FK_ParticipantsHtmlTemplate_TaskGroupTemplate ON shieldcoredb.task_group_template (HtmlMapperId);
CREATE INDEX FK_ScheduleType_TaskGroupTemplate_idx ON shieldcoredb.task_group_template (ScheduleTypeId);
CREATE INDEX task_group_template_ibfk_1_idx ON shieldcoredb.task_group_template (RelatedArchiveConnectionId);

-- Deployment Role Connection Template table
CREATE TABLE IF NOT EXISTS shieldcoredb.deployment_role_connection_template (
    Id                  INT AUTO_INCREMENT PRIMARY KEY,
    TaskGroupTemplateId INT NOT NULL,
    ServerId            INT NOT NULL,
    RoleId              INT NOT NULL,
    Port                INT NOT NULL,
    CONSTRAINT FK_DeploymentRoleConnectionTemplate_TaskGroupTemplate FOREIGN KEY (TaskGroupTemplateId) REFERENCES task_group_template (Id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT FK_DeploymentRoleConnectionTemplate_TgmRole FOREIGN KEY (RoleId) REFERENCES deployment_role (Id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE INDEX FK_DeploymentRoleConnectionTemplate_Server_idx ON shieldcoredb.deployment_role_connection_template (ServerId);
CREATE INDEX FK_DeploymentRoleConnectionTemplate_TgmRole_idx ON shieldcoredb.deployment_role_connection_template (TaskGroupTemplateId);
CREATE INDEX FK_DeploymentRoleConnectionTemplate_TgmRole_idx1 ON shieldcoredb.deployment_role_connection_template (RoleId);

-- Task Group table
CREATE TABLE IF NOT EXISTS shieldcoredb.task_group (
    Id                                     INT AUTO_INCREMENT PRIMARY KEY,
    TaskGroupTemplateId                    INT              NOT NULL,
    Name                                   VARCHAR(128)     NOT NULL,
    ScheduleTypeId                         INT              NOT NULL,
    StartTime                              DATETIME         NOT NULL,
    TimeInterval                           BIGINT           NOT NULL,
    IsRunImmediatelyAfterTaskGroupFinished BIT              NOT NULL,
    MdeGroupId                             INT              NOT NULL,
    IsMessageQueue                         BIT              NOT NULL,
    IsCompleteMissingTaskInRecovery        BIT              DEFAULT b'0' NOT NULL,
    PlatformName                           VARCHAR(128)     NOT NULL,
    EcommTypeId                            INT              NOT NULL,
    JsonMapperId                           INT              NOT NULL,
    RetentionDays                          INT              NOT NULL,
    IsCopyToArchive                        BIT              NOT NULL,
    RelatedArchiveConnectionId             INT              DEFAULT 0 NOT NULL,
    IsActive                               BIT              NOT NULL,
    DefaultLogLevelId                      INT              NOT NULL,
    EmlMapperId                            INT              DEFAULT 0 NULL,
    HtmlMapperId                           INT              NULL,
    IsHistoric                             BIT              DEFAULT b'0' NOT NULL,
    IsArchive                              BIT              DEFAULT b'1' NOT NULL,
    IsArchiveRetention                     BIT              DEFAULT b'1' NOT NULL,
    IsMigrationTaskGroup                   BIT              DEFAULT b'0' NOT NULL,
    CONSTRAINT FK_DrJsonTemplate_TaskGroup FOREIGN KEY (JsonMapperId) REFERENCES dr_json_template (Id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT FK_MdeGroup_TaskGroup FOREIGN KEY (MdeGroupId) REFERENCES mde_group (Id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT FK_ScheduleType_TaskGroup FOREIGN KEY (ScheduleTypeId) REFERENCES schedule_type (Id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT FK_TaskGroupTemplate_TaskGroup FOREIGN KEY (TaskGroupTemplateId) REFERENCES task_group_template (Id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT task_group_ibfk_1 FOREIGN KEY (EmlMapperId) REFERENCES eml_json_template (Id),
    CONSTRAINT task_group_ibfk_2 FOREIGN KEY (HtmlMapperId) REFERENCES participants_html_template (Id)
);

CREATE INDEX FK_DrJsonTemplate_TaskGroup_idx ON shieldcoredb.task_group (JsonMapperId);
CREATE INDEX FK_EcommType_TaskGroup_idx ON shieldcoredb.task_group (EcommTypeId);
CREATE INDEX FK_EmlJsonTemplate_TaskGroup ON shieldcoredb.task_group (EmlMapperId);
CREATE INDEX FK_LogLevel_TaskGroup_idx ON shieldcoredb.task_group (DefaultLogLevelId);
CREATE INDEX FK_MdeGroup_TaskGroup_idx ON shieldcoredb.task_group (MdeGroupId);
CREATE INDEX FK_ParticipantsHtmlTemplate_TaskGroup ON shieldcoredb.task_group (HtmlMapperId);
CREATE INDEX FK_ScheduleType_TaskGroup_idx ON shieldcoredb.task_group (ScheduleTypeId);
CREATE INDEX FK_ServerInfo_TaskGroup_idx ON shieldcoredb.task_group (RelatedArchiveConnectionId);
CREATE INDEX FK_TaskGroupTemplate_TaskGroup_idx ON shieldcoredb.task_group (TaskGroupTemplateId);

-- Add foreign key constraint to policy_base_retention now that task_group exists
ALTER TABLE shieldcoredb.policy_base_retention
ADD CONSTRAINT policy_base_retention_ibfk_3 FOREIGN KEY (TaskGroupId) REFERENCES task_group (Id) ON UPDATE CASCADE ON DELETE CASCADE;

-- Deployment Role Connection table
CREATE TABLE IF NOT EXISTS shieldcoredb.deployment_role_connection (
    Id          INT AUTO_INCREMENT PRIMARY KEY,
    TaskGroupId INT NOT NULL,
    ServerId    INT NOT NULL,
    RoleId      INT NOT NULL,
    Port        INT NOT NULL,
    CONSTRAINT FK_TaskGroupRoles_TaskGroup FOREIGN KEY (TaskGroupId) REFERENCES task_group (Id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT FK_TaskGroupRoles_TgmRole FOREIGN KEY (RoleId) REFERENCES deployment_role (Id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE INDEX FK_TaskGroupRoles_Server_idx ON shieldcoredb.deployment_role_connection (ServerId);
CREATE INDEX FK_TaskGroupRoles_TgmRole_idx ON shieldcoredb.deployment_role_connection (TaskGroupId);
CREATE INDEX FK_TaskGroupRoles_TgmRole_idx1 ON shieldcoredb.deployment_role_connection (RoleId);

-- Task Template table
CREATE TABLE IF NOT EXISTS shieldcoredb.task_template (
    Id                  INT AUTO_INCREMENT PRIMARY KEY,
    Name                VARCHAR(128)  NOT NULL,
    TypeId              INT           NOT NULL,
    StepNumber          INT           NOT NULL,
    CommandLine         VARCHAR(4096) NOT NULL,
    TaskGroupTemplateId INT           NOT NULL,
    PluginId            INT           NOT NULL,
    LogLevelId          INT           NOT NULL,
    IsActive            BIT           NOT NULL,
    CONSTRAINT FK_Plugin_TaskTemplate FOREIGN KEY (PluginId) REFERENCES plugin (Id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT FK_TaskGroupTemplate_TaskTemplate FOREIGN KEY (TaskGroupTemplateId) REFERENCES task_group_template (Id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT FK_TaskType_TaskTemplate FOREIGN KEY (TypeId) REFERENCES task_type (Id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE INDEX FK_LogLevel_TaskTemplate_idx ON shieldcoredb.task_template (LogLevelId);
CREATE INDEX FK_Plugin_TaskTemplate_idx ON shieldcoredb.task_template (PluginId);
CREATE INDEX FK_TaskGroupTemplate_TaskTemplate_idx ON shieldcoredb.task_template (TaskGroupTemplateId);
CREATE INDEX FK_TaskType_TaskTemplate_idx ON shieldcoredb.task_template (TypeId);

-- Task table
CREATE TABLE IF NOT EXISTS shieldcoredb.task (
    Id          INT AUTO_INCREMENT PRIMARY KEY,
    Name        VARCHAR(128)  NOT NULL,
    TypeId      INT           NOT NULL,
    StepNumber  INT           NOT NULL,
    CommandLine VARCHAR(4096) NOT NULL,
    TaskGroupId INT           NOT NULL,
    PluginId    INT           NOT NULL,
    LogLevelId  INT           NOT NULL,
    IsActive    BIT           NOT NULL,
    CONSTRAINT FK_Plugin_Task FOREIGN KEY (PluginId) REFERENCES plugin (Id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT FK_TaskGroup_Task FOREIGN KEY (TaskGroupId) REFERENCES task_group (Id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT FK_TaskType_Task FOREIGN KEY (TypeId) REFERENCES task_type (Id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE INDEX FK_LogLevel_Task_idx ON shieldcoredb.task (LogLevelId);
CREATE INDEX FK_Plugin_Task_idx ON shieldcoredb.task (PluginId);
CREATE INDEX FK_TaskGroup_Task_idx ON shieldcoredb.task (TaskGroupId);
CREATE INDEX FK_TaskType_Task_idx ON shieldcoredb.task (TypeId);

-- Parsing Step table
CREATE TABLE IF NOT EXISTS shieldcoredb.parsing_step (
    Id                  INT AUTO_INCREMENT PRIMARY KEY,
    TaskId              INT          NOT NULL,
    TaskGroupTypeId     INT          NOT NULL,
    StepNumber          INT          NOT NULL,
    Name                VARCHAR(512) NOT NULL,
    Xpath               VARCHAR(512) NOT NULL,
    ParentIds           VARCHAR(128) NOT NULL,
    IsAttribute         BIT          NOT NULL,
    IsAttributeOnEntity BIT          NOT NULL,
    IsGlobal            BIT          NOT NULL,
    IsActive            BIT          NOT NULL,
    CONSTRAINT FK_Task_ParsingStep FOREIGN KEY (TaskId) REFERENCES task (Id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE INDEX FK_TaskGroupType_ParsingStep_idx ON shieldcoredb.parsing_step (TaskGroupTypeId);
CREATE INDEX TaskId_idx ON shieldcoredb.parsing_step (TaskId);

-- Plugin Store table
CREATE TABLE IF NOT EXISTS shieldcoredb.tbl_plugin_store (
    Id                 INT          NOT NULL AUTO_INCREMENT PRIMARY KEY,
    PluginId           INT          NOT NULL,
    TaskGroupId        INT          NOT NULL,
    InputKey           VARCHAR(128) NOT NULL,
    InputValue         VARCHAR(128) NOT NULL,
    InputValuePrevious VARCHAR(128) NOT NULL,
    CONSTRAINT FK_Plugin_TblPluginStore FOREIGN KEY (PluginId) REFERENCES plugin (Id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT FK_TaskGroup_TblPluginStore FOREIGN KEY (TaskGroupId) REFERENCES task_group (Id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE INDEX FK_Plugin_TblPluginStore_idx ON shieldcoredb.tbl_plugin_store (PluginId);
CREATE INDEX FK_TaskGroup_TblPluginStore_idx ON shieldcoredb.tbl_plugin_store (TaskGroupId);

-- Search Participant Mapping table
CREATE TABLE IF NOT EXISTS shieldcoredb.search_participant_mapping (
    id                      INT AUTO_INCREMENT PRIMARY KEY,
    email_column_names      JSON NOT NULL COMMENT 'Array of email column names from users table',
    to_participant_email_key   VARCHAR(255) NOT NULL DEFAULT 'participantEmail',
    from_participant_email_key VARCHAR(255) NOT NULL DEFAULT 'participantEmail',
    cc_participant_email_key   VARCHAR(255) NOT NULL DEFAULT 'participantEmail',
    bcc_participant_email_key  VARCHAR(255) NOT NULL DEFAULT 'participantEmail',
    created_at             TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at             TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- ============================================================================
-- END OF SCRIPT
-- ============================================================================
-- Notes:
-- 1. The 'users' table in activedirectoryusers schema is created dynamically
--    by the DSS service via the HR data API endpoint (/dss/v3.9/hrmetadata)
-- 2. Stored procedures (select_users, select_users_email) are also created
--    dynamically by the service based on HR field configuration
-- 3. This script creates the base structure. Additional data (plugins, MDEs,
--    templates) should be populated via the DSS API endpoints
-- ============================================================================
