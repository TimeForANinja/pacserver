# Use Apache as base image
FROM httpd:latest

# Copy your demo files
COPY demo_files ./demo_files

# Replace {{ .Contact }} with <!--#echo var="contactVAR"--> in all files
RUN find ./demo_files -type f -exec sed -i 's/{{ .Contact }}/<!--#echo var="contactVAR"-->/g' {} +

# update config
COPY <<'EOF' /usr/local/apache2/conf/httpd.conf
# Basic server settings
ServerRoot "/usr/local/apache2"
Listen 8080
ServerAdmin you@example.com
ServerName localhost

# Load essential modules
LoadModule mpm_event_module modules/mod_mpm_event.so
LoadModule unixd_module modules/mod_unixd.so
LoadModule authz_core_module modules/mod_authz_core.so
LoadModule dir_module modules/mod_dir.so
LoadModule mime_module modules/mod_mime.so
LoadModule rewrite_module modules/mod_rewrite.so
LoadModule log_config_module modules/mod_log_config.so
LoadModule include_module modules/mod_include.so
LoadModule negotiation_module modules/mod_negotiation.so

# MIME configuration
TypesConfig conf/mime.types
AddType application/x-ns-proxy-autoconfig .pac
# enable SSI (Server Side Includes) part 1
AddHandler server-parsed .pac

# Basic directory structure
DocumentRoot "/usr/local/apache2/demo_files"

# Directory permissions
<Directory "/usr/local/apache2/demo_files">
    Options Indexes FollowSymLinks
    #AllowOverride None
    AllowOverride All
    Require all granted
    # enable SSI (Server Side Includes) part 2
    Options +Includes
</Directory>


# Rewrite rules
RewriteEngine On
RewriteMap lowercase int:tolower


# Define Variables
RewriteRule .* - [E=contactDEFAULT:Your-Service-Desk]
RewriteRule .* - [E=contactHEL:Helsinki-Servicedesk]
RewriteRule .* - [E=contactFRA:Frankfurt-Servicedesk]
RewriteRule .* - [E=contactPAR:Paris-Servicedesk]
RewriteRule .* - [E=contactLON:London-Servicedesk]
RewriteRule .* - [E=contactUSA:America-Servicedesk]


# Get IP from URL
RewriteCond %{REQUEST_URI} ^\/([0-9].*)$
RewriteRule ^/.*$   /%1                     [C]
RewriteRule ^/(.*)$ -                       [E=addr:$1,T=application/x-ns-proxy-autoconfig,S=1]

# Read in Client Address and store it in %{addr}
RewriteCond %{REMOTE_ADDR} ^(.*)$
RewriteRule ^/.*$   /%1                     [C]
RewriteRule ^/(.*)$ -                       [E=addr:$1,T=application/x-ns-proxy-autoconfig]


# overwrite the country zones
RewriteCond %{ENV:addr}    ^172.16(\.|$)                                       [OR]
RewriteCond %{ENV:addr}    ^172.16.[1-2](\.|$)
RewriteRule ^/.*$ /usr/local/apache2/demo_files/pacs/countries/usa.pac         [L,E=contactVAR:%{env:contactUSA}]

RewriteCond %{ENV:addr}    ^172.17(\.|$)                                       [OR]
RewriteCond %{ENV:addr}    ^172.17.[1-2](\.|$)
RewriteRule ^/.*$ /usr/local/apache2/demo_files/pacs/countries/uk.pac          [L,E=contactVAR:%{env:contactLON}]

RewriteCond %{ENV:addr}    ^172.18(\.|$)                                       [OR]
RewriteCond %{ENV:addr}    ^172.18.1(\.|$)
RewriteRule ^/.*$ /usr/local/apache2/demo_files/pacs/countries/france.pac      [L,E=contactVAR:%{env:contactPAR}]

RewriteCond %{ENV:addr}    ^172.19(\.|$)                                       [OR]
RewriteCond %{ENV:addr}    ^172.19.1(\.|$)
RewriteRule ^/.*$ /usr/local/apache2/demo_files/pacs/countries/japan.pac       [L,E=contactVAR:%{env:contactDEFAULT}]

RewriteCond %{ENV:addr}    ^172.20(\.|$)                                       [OR]
RewriteCond %{ENV:addr}    ^172.20.1(\.|$)
RewriteRule ^/.*$ /usr/local/apache2/demo_files/pacs/countries/australia.pac   [L,E=contactVAR:%{env:contactDEFAULT}]

RewriteCond %{ENV:addr}    ^172.21(\.|$)                                       [OR]
RewriteCond %{ENV:addr}    ^172.21.1(\.|$)
RewriteRule ^/.*$ /usr/local/apache2/demo_files/pacs/countries/canada.pac      [L,E=contactVAR:%{env:contactUSA}]

RewriteCond %{ENV:addr}    ^172.22(\.|$)                                      [OR]
RewriteCond %{ENV:addr}    ^172.22.1(\.|$)
RewriteRule ^/.*$ /usr/local/apache2/demo_files/pacs/countries/brazil.pac      [L,E=contactVAR:%{env:contactDEFAULT}]

RewriteCond %{ENV:addr}    ^172.23(\.|$)                                       [OR]
RewriteCond %{ENV:addr}    ^172.23.1(\.|$)
RewriteRule ^/.*$ /usr/local/apache2/demo_files/pacs/countries/india.pac       [L,E=contactVAR:%{env:contactDEFAULT}]

RewriteCond %{ENV:addr}    ^172.24(\.|$)                                       [OR]
RewriteCond %{ENV:addr}    ^172.24.1(\.|$)
RewriteRule ^/.*$ /usr/local/apache2/demo_files/pacs/countries/china.pac       [L,E=contactVAR:%{env:contactDEFAULT}]

RewriteCond %{ENV:addr}    ^172.25(\.|$)                                       [OR]
RewriteCond %{ENV:addr}    ^172.25.1(\.|$)
RewriteRule ^/.*$ /usr/local/apache2/demo_files/pacs/countries/russia.pac      [L,E=contactVAR:%{env:contactDEFAULT}]


# Fallback PAC
RewriteRule ^/.*$ /usr/local/apache2/demo_files/pacs/my-company.pac            [L,E=contactVAR:%{contactDEFAULT}]


# Basic logging
ErrorLog /proc/self/fd/2
LogLevel warn
LogFormat "%h %l %u %t \"%r\" %>s %b" common
CustomLog /proc/self/fd/1 common
EOF

# Expose port for Apache
EXPOSE 8080

# Start Apache in foreground
ENTRYPOINT ["httpd-foreground"]
