-- https://stackoverflow.com/a/49058059/88646
-- CC BY-SA 3.0
function Link (link)
  link.target = link.target:gsub('.md$', '.html')
  return link
end