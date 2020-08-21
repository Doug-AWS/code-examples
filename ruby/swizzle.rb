# Get input string from command line
if ARGV.length == 0
    puts 'You must supply a string argument'
    return
end

chars = ARGV[0].split('')
output = ""

chars.each { |c|
    if ((c.ord >= 'a'.ord) && (c.ord <= 'z'.ord))
        output += c.upcase
    elsif ((c.ord >= 'A'.ord) && (c.ord <= 'Z'.ord))
        output += c.downcase
    else
        output += c
    end        
}

puts output