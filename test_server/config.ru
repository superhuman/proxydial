require 'bundler'
Bundler.require

get '/' do
    "moo"
end

get '/local' do
    redirect 'http://127.0.0.1/'
end

get '/v6' do
    redirect 'http://[::1]/'
end

get '/remote' do
    redirect '/nowhere'
end

get '/nowhere' do
    'DONE'
end

get '/file' do
    redirect 'file:///etc/passwd'
end

get '/recurse' do
    redirect '/recurse?diff=' + rand.to_s
end

get '/port' do
    redirect 'https://proxydial.herokuapp.com:25/'
end

run Sinatra::Application

