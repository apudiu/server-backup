### What

This is a utility to perform websites backup from remote servers.
But this can be used to backup any file/ directory from any machine. Assuming you've right permissions to do that over
SSH.

### Why

Just for internal use-case. We've (in my work) lots of websites in lots of servers. I'm responsible to keep backups of
those as per requirement.
Unfortunately we've no dev-ops person for that.
<br >
So I've written this utility which will do automated backups of those websites. If anybody are in similar situation can
use this tool :)

### How

1. Download appropriate binary from **Releases**. for ex: `server-backup-linux-amd64` (calling it `bin`)
2. Execute `bin gen` to generate sample configuration.
3. Customize parameters in `./config/servers.yml` & `./config/[server-ip]/[project-name].yml` with your data
4. Run backup by executing `./bin`

#4 can be added in cron for automated execution. So this can trigger automatic backups at desired intervals.

### Features

1. Backup project files as zip
2. Can specify ignore list in the zip
3. Export database (currently MySQL/ MariaDB is supported) as zip
4. Transfer files & DB backup zip in local
5. Upload this in S3
6. Keep specified number of backups for each project (website) from each server
7. Project backup logs are included in backup directory (uploaded in S3 too)
8. There's `[backup-dir]/run.log` where a log summery is available (this is not available in S3)

#### Config parameters `./config/servers.yml`
The configuration works like this. First you define a server and list which directories should be backed up. <br> 
*This file can contain config for multiple servers. Following is example for one server*
<table>
    <thead>
    <tr>
        <th>Key</th>
        <th>Required?</th>
        <th>Description</th>
    </tr>
    </thead>
    <tbody>
    <tr>
        <td>privateKeyPath</td>
        <td>y</td>
        <td>
            If you've a private key (PK) for the server specify it's location (in local fs)
        </td>
    </tr>
    <tr>
        <td>ip</td>
        <td>y</td>
        <td>Server IP address (where your websites are deployed)</td>
    </tr>
    <tr>
        <td>port</td>
        <td>y</td>
        <td>SSH port</td>
    </tr>
    <tr>
        <td>user</td>
        <td>y</td>
        <td>
            username who can:<br/>
            1. SSH into the server <br/> 
            2. Access those directories that need to be backed up<br>
            3. Write to those directories where backup files will be saved
        </td>
    </tr>
    <tr>
        <td>password</td>
        <td>y</td>
        <td>
            Password of the provided user. if you don't have PK or the PK is password protected, you need to specify the password. <br>
            * if the PK is password protected this will be used to parse the PK <br>
            * if no key provided this password will be used to log in as password auth
        </td>
    </tr>
    <tr>
        <td>projectRoot</td>
        <td>y</td>
        <td>
            Working directory in the server where projects are located. Projects must be under this directory.
        </td>
    </tr>
    <tr>
        <td>backupSources</td>
        <td>y</td>
        <td>
            List directories (project/ websites). This must be immediate child of <strong>projectRoot</strong>.Only specified projects will be backed. <br> 
            If you specify a project here, corresponding project config should reside in <code>./config/[ip]/[backupSource[n]].yml</code>. <br><br>
            For ex: if your server ip is <code>192.168.16.18</code> and your website folder is <code>foo-website</code>. Then <code>foo-website</code> should be listed in this config and project config should reside in <code>./config/192.168.16.18/foo-website.yml</code>. You need to create it by copying existing one or customizing generated one
        </td>
    </tr>
    <tr>
        <td>backupDestPath</td>
        <td>n</td>
        <td>
            File backups will be placed in this path. <br>
            Specify custom backup path in local fs (ff you'd like), If not specified this will use <code>./backups</code> as local destination
        </td>
    </tr>
    <tr>
        <td>s3User</td>
        <td>n</td>
        <td>
            If you need to transfer your backups in AWS S3 then you need to specify it. <br>
            Locally configured AWS credential profile name (this is not a IAM user's name) <br>
            <i>Specified profile should've appropriate permission to the bucket <strong>s3Bucket</strong></i>
        </td>
    </tr>
    <tr>
        <td>s3Bucket</td>
        <td>n</td>
        <td>
            Required if you specified <strong>s3User</strong> <br>
            AWS S3 bucket name where the provided user has rw permission
        </td>
    </tr>
    </tbody>
</table>

Following is an example of `servers.yml`

```yml
servers:
  # If you've a private key (PK) for the server specify it's location
  - privateKeyPath: /home/user/serverKey.pem
    # server ip address for ssh
    ip: 192.168.0.100
    # ssh port
    port: 22
    # ssh user
    user: privilegedUserWhoCanDoYourTasks
    # Provide password if you don't have PK or the PK is password protected
    # if the PK is password protected this password will be used to parse the PK
    # if no key provided this password will be used to log in as password auth
    password: "123456"
    # working directory in the server where projects are located
    # projects must be under this dir
    projectRoot: /var/www/php80
    # list project (dir) names under @projectRoot
    # only specified projects will be backed up
    # if you specify a project here, corresponding project config should reside in @ip/<project name>.yml
    backupSources:
      - order-online
      - buy-sell
    # If you like specify custom backup path
    backupDestPath: ""
    # AWS user who has appropriate permissions for uploading files
    # this user/ profile need to be configured in accessible way in the runner machine
    s3User: s3-user-who-can-upload-to-the-bucket
    # AWS S3 bucket name where the provided user can upload files
    s3Bucket: s3-bucket-name
```

You can find this in `./config_sample` directory or can generate sample one in above mentioned way.

#### Config parameters for project `./config/[server-ip]/[project-dir].yml`

*This file can contain config for one project/ website*
<table>
    <thead>
        <tr>
            <th>Key</th>
            <th>Required?</th>
            <th>Description</th>
        </tr>
    </thead>
    <tbody>
        <tr>
            <td>path</td>
            <td>y</td>
            <td>
                Directory name (inside server.projectRoot). <br>
                This directory will be zipped & downloaded
            </td>
        </tr>
        <tr>
            <td>excludePaths</td>
            <td>n</td>
            <td>
                List of paths to exclude while zipping. <br>
                * If you'd like to exclude whole directory you should do it like <code>dir/to/exclude/*</code>
            </td>
        </tr>
        <tr>
            <td>zipFileName</td>
            <td>n</td>
            <td>
                Customize zip name if needed. By default this will be like yyyy_mm_dd_@path.zip. It's recommended not to change this unless necessary.
            </td>
        </tr>
        <tr>
            <td><strong>envFileInfo</strong></td>
            <td>n</td>
            <td>
                This section is required when you want to backup your DB and don't have / want to provide DB credentials in explicitly in <strong>dbInfo</strong> section. <br>
                This section is used to parse provided env file to backup your DB (currently MySQL/ MariaDB) is supported). <br>
                If you do not need DB backup just leave this empty or delete this section
            </td>
        </tr>
        <tr>
            <td><strong>envFileInfo</strong>.path</td>
            <td>n</td>
            <td>
                Provide path of a .env file inside project <strong>path</strong>. <br>
                Contents of this file will be parsed by provided keys & used to dump the db
            </td>
        </tr>
        <tr>
            <td><strong>envFileInfo</strong>.dbHostKeyName</td>
            <td>n</td>
            <td>
                Inside `.env` file, the key name that holds value for DB host IP address 
            </td>
        </tr>
        <tr>
            <td><strong>envFileInfo</strong>.dbPortKeyName</td>
            <td>n</td>
            <td>
                Inside `.env` file, the key name that holds value for DB port number 
            </td>
        </tr>
        <tr>
            <td><strong>envFileInfo</strong>.dbUserKeyName</td>
            <td>n</td>
            <td>
                Inside `.env` file, the key name that holds value for DB username 
            </td>
        </tr>
        <tr>
            <td><strong>envFileInfo</strong>.dbPassKeyName</td>
            <td>n</td>
            <td>
                Inside `.env` file, the key name that holds value for DB password for the user (<strong>envFileInfo.dbUserKeyName</strong>) 
            </td>
        </tr>
        <tr>
            <td><strong>envFileInfo</strong>.dbNameKeyName</td>
            <td>n</td>
            <td>
                Inside `.env` file, the key name that holds value for DB name 
            </td>
        </tr>
        <tr>
            <td><strong>dbInfo</strong></td>
            <td>n</td>
            <td>
                This section is required when you want to backup your DB and don't have <code>.env</code> file or don't want to provide it in <strong>envFileInfo</strong> section. <br>
                This section is used to login to database for dumping. <br>
                You can provide both the <strong>envFileInfo</strong> and this section so if env parsing fails then this values will be used. When both section provided, (parsed)value from <strong>envFileInfo</strong> section will override this section values. 
            </td>
        </tr>
        <tr>
            <td><strong>dbInfo</strong>.hostIp</td>
            <td>n</td>
            <td>
                DB host IP address 
            </td>
        </tr>
        <tr>
            <td><strong>dbInfo</strong>.port</td>
            <td>n</td>
            <td>
                DB port number 
            </td>
        </tr>
        <tr>
            <td><strong>dbInfo</strong>.user</td>
            <td>n</td>
            <td>
                DB username 
            </td>
        </tr>
        <tr>
            <td><strong>dbInfo</strong>.pass</td>
            <td>n</td>
            <td>
                DB <strong>dbInfo.user</strong>'s password 
            </td>
        </tr>
        <tr>
            <td><strong>dbInfo</strong>.name</td>
            <td>n</td>
            <td>
                DB name 
            </td>
        </tr>
        <tr>
            <td>backupCopies</td>
            <td>n</td>
            <td>
                Number of backup copies to keep, if not specified or 0 (zero) is provided then by default 3 latest copies of backup will be kept and rest will be deleted.
                It'll keep provided number of copies in local & S3 (if provided). <br> <br>
                <i>For Ex: if you specify 5, to keep latest 5 copies of this project then this will backup first and then check if there's more than 5 copies in local & S3, If any extra copy is found, it'll delete that (form local & S3)</i>
            </td>
        </tr>


    </tbody>
</table>

Following is an example of `./config/[server-ip]/[project-dir].yml`

```yml
# project path inside server.projectRoot
path: order-online
# exclude this paths when zipping
excludePaths:
  # for ignoring whole directory it must end with "/*"
  - api/vendor/*
  - api/storage/framework/*
  - api/storage/logs/*
  - api/.rsyncIgnore
  - www/vendor/*
# customize zip name if needed
# by default this will be like yyyy_mm_dd_@path.zip
zipFileName: ""
# For db backup
envFileInfo:
  # Provide path of a .env file,
  # contents of this fill will be parsed by provided keys & used to dump the db
  path: api/.env
  dbHostKeyName: DB_HOST
  dbPortKeyName: DB_PORT
  dbUserKeyName: DB_USERNAME
  dbPassKeyName: DB_PASSWORD
  dbNameKeyName: DB_DATABASE
# if .env is not provided, provide info directly
# when env file path provided that will be used instead values provided here
dbInfo:
  hostIp: ""
  port: 0
  user: ""
  pass: ""
  name: ""
# number of backup copies to keep, if not specified of 0 is provided
# then by default 3 latest copies of backup will be kept & rest will be deleted
backupCopies: 5
```

You can find this in `./config/[server-ip]/[project-dir].yml` directory or can generate sample one in above mentioned way.

### Contribution
I've built what we need till now. But I think there's possibility that this tool can be a great too which works with many different DB's. Supports hostnames instead of server ip's etc.
If anybody need this kind of features, go ahead and extend/ modify this. If you think you'r work can be useful & improve this tool then submit a PR. I'm open to appreciate that.

### Getting Help
Something not working for you? Open an issue, I'll try to help :)
